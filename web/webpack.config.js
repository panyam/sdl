const fs = require("fs");
const path = require("path");
const webpack = require("webpack");
const CopyPlugin = require("copy-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const HtmlWebpackTagsPlugin = require("html-webpack-tags-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const { CleanWebpackPlugin } = require("clean-webpack-plugin");
// const uglifyJsPlugin = require("uglifyjs-webpack-plugin");

const TEMPLATES_FOLDER = "./templates";

// Read Samples first
function readdir(path) {
  const items = fs.readdirSync(path);
  return items.map(function (item) {
    let file = path;
    if (item.startsWith("/") || file.endsWith("/")) {
      file += item;
    } else {
      file += "/" + item;
    }
    const stats = fs.statSync(file);
    return { file: file, name: item, stats: stats };
  });
}

const components = [
  // "NotationViewer",
  // "NotationEditor",
  //"ConsoleView",
  "CaseStudyPage"
];

module.exports = (_env, options) => {
  context: path.resolve(__dirname);
  const isDevelopment = options.mode == "development";
  const webpackConfigs = {
    devtool: "source-map",
    devServer: {
      hot: true,
      serveIndex: true,
      // contentBase: path.join(__dirname, "../dist/static/dist"),
      before: function (app, server) {
        app.get(/\/dir\/.*/, function (req, res) {
          const path = "./" + req.path.substr(5);
          console.log("Listing dir: ", path);
          const listing = readdir(path);
          res.json({ entries: listing });
        });
      },
    },
    externals: {
      // CodeMirror: 'CodeMirror',
      // 'GL': "GoldenLayout",
      ace: "commonjs ace",
      // ace: 'ace',
    },
    optimization: {
      splitChunks: {
        chunks: "all",
      },
    },
    output: {
      path: path.resolve(__dirname, "./static/js/gen/"),
      publicPath: "/static/js/gen/",
      // filename: "[name]-[hash:8].js",
      filename: "[name].[contenthash].js",
      library: ["notation", "[name]"],
      libraryTarget: "umd",
      umdNamedDefine: true,
      globalObject: "this",
    },
    module: {
      rules: [
        {
          test: /\.js$/,
          exclude: ["node_modules/", "dist"].map((x) => path.resolve(__dirname, x)),
          use: ["babel-loader"],
        },
        {
          test: /\.ts$/,
          exclude: [path.resolve(__dirname, "node_modules"), path.resolve(__dirname, "dist")],
          include: [`${TEMPLATES_FOLDER}/`].map((x) => path.resolve(__dirname, x)),
          use: [
            {
              loader: "ts-loader",
              options: { configFile: "tsconfig.json" },
            },
          ],
        },
        {
          test: /\.s(a|c)ss$/,
          use: [
            MiniCssExtractPlugin.loader,
            // "style-loader",
            "css-loader",
            {
              loader: "sass-loader",
              options: {
                sourceMap: isDevelopment,
              },
            },
          ],
        },
        {
          test: /\.(png|svg|jpg|jpeg|gif)$/i,
          type: "asset/resource",
        },
        {
          test: /\.(woff|woff2|eot|ttf|otf)$/i,
          type: "asset/resource",
        },
      ],
    },
    entry: components.reduce(function (map, comp) {
      map[comp] = path.join(__dirname, `${TEMPLATES_FOLDER}/${comp}.ts`);
      return map;
    }, {}),
    plugins: [
      new CleanWebpackPlugin(),
      new MiniCssExtractPlugin(),
      ...components.map(
        (component) =>
          new HtmlWebpackPlugin({
            chunks: [component],
            // inject: false,
            filename: path.resolve(__dirname, `${TEMPLATES_FOLDER}/gen.${component}.html`),
            // template: path.resolve(__dirname, `${component}.html`),
            templateContent: "",
            minify: { collapseWhitespace: false },
          }),
      ),
      new webpack.HotModuleReplacementPlugin(),
    ],
    resolve: {
      extensions: [".js", ".jsx", ".ts", ".tsx", ".scss", ".css", ".png"],
      fallback: {
        "crypto-browserify": require.resolve("crypto-browserify"), //if you want to use this module also don't forget npm i crypto-browserify
        "querystring-es3": false,
        assert: false,
        buffer: false,
        child_process: false,
        crypto: false,
        fs: false,
        http: false,
        https: false,
        net: false,
        os: false,
        path: false,
        querystring: false,
        stream: false,
        tls: false,
        url: false,
        util: false,
        zlib: false,
      },
    },
  };
  if (false && !isDevelopment) {
    webpackConfigs.plugins.splice(0, 0, new uglifyJsPlugin());
  }
  return webpackConfigs;
};
