const webpack = require('webpack'),
      CopyWebpackPlugin = require('copy-webpack-plugin'),
      autoprefixer = require('autoprefixer'),
      path = require('path');

const NITROUS_ENV = process.env.NITROUS_ENV || 'development';

const SRC_DIR = path.resolve(__dirname, 'ide'),
      BUILD_DIR = path.resolve(__dirname, 'build'),
      NODE_MODULES_DIR = path.resolve(__dirname, 'node_modules'),
      APP_DIR = path.join(SRC_DIR, 'app'),
      DEST_DIR = path.join(BUILD_DIR, NITROUS_ENV),
      ASSETS_DIR = path.join(DEST_DIR, 'assets');

const config = {
  entry: path.join(APP_DIR, 'main.js'),
  output: {
    path: ASSETS_DIR,
    filename: 'bundle.js'
  },
  devtool: "source-map",
  plugins: [
    new CopyWebpackPlugin([
      {
        from: path.join(SRC_DIR, 'assets'),
        to: ASSETS_DIR
      }
    ])
  ],
  module: {
    loaders: [
      {
        test: /\.jsx?/,
        include: APP_DIR,
        loaders: ['babel', 'eslint'],
      },
      {
        test: /\.css$/,
        include: [APP_DIR, NODE_MODULES_DIR],
        loaders: ["style", "css", "postcss"]
      },
      {
        test: /\.s(a|c)ss$/,
        include: [APP_DIR, NODE_MODULES_DIR],
        loaders: ["style", "css", "postcss", "sass"]
      }
    ]
  },
  postcss() {
    return [autoprefixer];
  },
};

module.exports = config;
