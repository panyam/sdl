// webpack.config.js

const fs = require("fs");
const path = require("path");
const webpack = require("webpack");
const CopyPlugin = require("copy-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");

const SRC_FOLDERS = ["./components"];
const OUTPUT_FOLDERS = ["./templates"]; // Where gen.*.html files go

const components = [
  ["CaseStudyPage", 0, "tsx"]
];

module.exports = (_env, options) => {
  const context = path.resolve(__dirname); // Project root context
  const isDevelopment = options.mode == "development";
  // Define output path for bundled JS and copied assets
  const outputDir = path.resolve(__dirname, "./static/js/gen/");
  // Define the public base path for the static directory (as served by the external server)
  const staticPublicPath = '/static'; // Assuming './static' is served at the root path '/static'

  return {
    context: context,
    devtool: "source-map",
    devServer: {
      hot: true,
      serveIndex: true,
    },
    externals: {
      ace: "commonjs ace",
    },
    entry: components.reduce(function (map, comp) {
      const compName = comp[0];
      const compFolder = SRC_FOLDERS[comp[1]];
      const compExt = comp[2];
      map[compName] = path.join(context, `${compFolder}/${compName}.${compExt}`);
      return map;
    }, {}),
    module: {
      rules: [
        {
          test: /\.jsx$/,
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
          exclude: path.resolve(context, "node_modules/"),
          use: ["babel-loader"],
        },
        /*
        {
          test: /\.tsx$/,
          use: 'ts-loader',
          exclude: /node_modules/,
        },
        */
        {
          test: /\.tsx?$/,
          exclude: path.resolve(context, "node_modules/"),
          include: SRC_FOLDERS.map((x) => path.resolve(context, x)),
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
           generator: {
                filename: 'assets/[hash][ext][query]' // Place assets in static/js/gen/assets/
           }
        },
        {
          test: /\.(woff|woff2|eot|ttf|otf)$/i,
          type: "asset/resource",
           generator: {
                 filename: 'assets/[hash][ext][query]' // Place assets in static/js/gen/assets/
           }
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
        "querystring-es3": false,
        assert: false,
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
        buffer: false,
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
      path: outputDir, // -> ./static/js/gen/
      // Public path where browser requests bundles/assets. Matches path structure served by static server.
      publicPath: `${staticPublicPath}/js/gen/`, // -> /static/js/gen/
      filename: "[name].[contenthash].js",
      library: ["leetcoach", "[name]"],
      libraryTarget: "umd",
      umdNamedDefine: true,
      globalObject: "this",
      clean: true, // Clean the output directory before build
    },
    plugins: [
      new webpack.ProvidePlugin({
        process: 'process/browser',
        React: 'react'
      }),
      new MiniCssExtractPlugin(),
      // These HTML files might be unnecessary if your server templating handles includes differently
      ...components.map(
        (component) =>
          new HtmlWebpackPlugin({
            chunks: [component[0]],
            filename: path.resolve(__dirname, `${OUTPUT_FOLDERS[component[1]]}/gen.${component[0]}.html`),
            templateContent: "",
            minify: false, // { collapseWhitespace: false },
            inject: 'body',
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
