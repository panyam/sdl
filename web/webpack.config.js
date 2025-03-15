const fs = require("fs");
const path = require("path");
const webpack = require("webpack");
const CopyPlugin = require("copy-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const HtmlWebpackTagsPlugin = require("html-webpack-tags-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const { CleanWebpackPlugin } = require("clean-webpack-plugin");
// const uglifyJsPlugin = require("uglifyjs-webpack-plugin");

const SRC_FOLDER = "./components"
const TEMPLATES_FOLDER = "./templates";

const components = [
  "CaseStudyPage"
];

module.exports = (_env, options) => {
  context: path.resolve(__dirname);
  const isDevelopment = options.mode == "development";
  return webpackConfigs = {
    devtool: "source-map",
    devServer: {
      hot: true,
      serveIndex: true,
    },
    externals: {
      // CodeMirror: 'CodeMirror',
      // 'GL': "GoldenLayout",
      ace: "commonjs ace",
      // ace: 'ace',
    },
    entry: components.reduce(function (map, comp) {
      map[comp] = path.join(__dirname, `${SRC_FOLDER}/${comp}.tsx`);
      return map;
    }, {}),
    module: {
      rules: [
      {
        test: /\.tsx?$/,
        use: 'ts-loader',
        exclude: /node_modules/,
      },
      {
        test: /\.jsx?$/,
        exclude: /node_modules/,
        use: {
          loader: 'ts-loader',
          options: {
            transpileOnly: true
          }
        },
      },
        {
          test: /\.js$/,
          exclude: ["node_modules/", "dist"].map((x) => path.resolve(__dirname, x)),
          use: ["babel-loader"],
        },
        {
          test: /\.ts$/,
          exclude: [path.resolve(__dirname, "node_modules"), path.resolve(__dirname, "dist")],
          include: [`${SRC_FOLDER}/`].map((x) => path.resolve(__dirname, x)),
          use: [
            {
              loader: "ts-loader",
              options: { configFile: "tsconfig.json" },
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
    resolve: {
      alias: {
        'react': path.resolve('./node_modules/react'),
        'react-dom': path.resolve('./node_modules/react-dom'),
      },
      extensions: [".js", ".jsx", ".ts", ".tsx", ".scss", ".css", ".png"],
      fallback: {
        /*
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
        */
        // Needed for Excalidraw
        "process": require.resolve("process/browser")
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
    plugins: [
      new webpack.ProvidePlugin({
        process: 'process/browser',
        React: 'react'
      }),
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
    optimization: {
      splitChunks: {
        chunks: "all",
      },
    },
  };
};
