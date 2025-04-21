// webpack.config.js

const fs = require("fs");
const path = require("path");
const webpack = require("webpack");
const CopyPlugin = require("copy-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");

const SRC_FOLDERS = ["./views/components"];
const OUTPUT_FOLDERS = ["./views/templates"]; // Where gen.*.html files go
const OUTPUT_DIR = path.resolve(__dirname, "./static/js/gen/");

const components = [
  // ["CaseStudyPage", 0, "tsx"]
  ["DetailPage", 0, "ts"],
  ["HomePage", 0, "ts"],
  ["LoginPage", 0, "ts"],
];

module.exports = (_env, options) => {
  const context = path.resolve(__dirname); // Project root context
  const isDevelopment = options.mode == "development";
  // Define output path for bundled JS and copied assets
  // Define the public base path for the static directory (as served by the external server)
  const staticPublicPath = '/static'; // Assuming './static' is served at the root path '/static'

  return {
    context: context,
    devtool: "source-map",
    // NO devServer block needed
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
      extensions: [".js", ".jsx", ".ts", ".tsx", ".css", ".png"],
      fallback: {
        // Needed for Excalidraw
        "process": require.resolve("process/browser"),
        "stream": require.resolve("stream-browserify"),
        "buffer": require.resolve("buffer"),
        "fs": false, "path": false, "os": false, "crypto": false, "http": false,
        "https": false, "net": false, "tls": false, "zlib": false, "url": false,
        "assert": false, "util": false, "querystring": false, "child_process": false
      },
    },
    output: {
      path: OUTPUT_DIR, // -> ./static/js/gen/
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

      // --- Copy TinyMCE Assets ---
      new CopyPlugin({
          patterns: [
              // Ensure paths are correct: Copy FROM node_modules TO the output directory
              {
                  from: path.resolve(context, 'node_modules/tinymce/skins'),
                  to: path.resolve(OUTPUT_DIR, 'skins'), // -> ./static/js/gen/skins
                  globOptions: { ignore: ['**/.*'] } // Ignore hidden files like .DS_Store
              },
              {
                  from: path.resolve(context, 'node_modules/tinymce/plugins'),
                  to: path.resolve(OUTPUT_DIR, 'plugins'), // -> ./static/js/gen/plugins
                  globOptions: { ignore: ['**/.*'] }
              },
              {
                  from: path.resolve(context, 'node_modules/tinymce/themes'),
                  to: path.resolve(OUTPUT_DIR, 'themes'), // -> ./static/js/gen/themes
                  globOptions: { ignore: ['**/.*'] }
              },
              {
                  from: path.resolve(context, 'node_modules/tinymce/icons'),
                  to: path.resolve(OUTPUT_DIR, 'icons'), // -> ./static/js/gen/icons
                  globOptions: { ignore: ['**/.*'] }
              },
              {
                  from: path.resolve(context, 'node_modules/tinymce/models'),
                  to: path.resolve(OUTPUT_DIR, 'models'), // -> ./static/js/gen/models
                  globOptions: { ignore: ['**/.*'] }
              },
          ],
      }),
    ],
    optimization: {
      splitChunks: {
        chunks: "all",
      },
    },
  };
};
