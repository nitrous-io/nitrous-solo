const webpack = require('webpack'),
      CopyWebpackPlugin = require('copy-webpack-plugin'),
      path = require('path'),
      ExtractTextPlugin = require('extract-text-webpack-plugin'),
      OptimizeCssAssetsPlugin = require('optimize-css-assets-webpack-plugin');

const NITROUS_ENV = process.env.NITROUS_ENV || 'development';

const SRC_DIR = path.resolve(__dirname, 'ide'),
      BUILD_DIR = path.resolve(__dirname, 'build'),
      NODE_MODULES_DIR = path.resolve(__dirname, 'node_modules'),
      APP_DIR = path.join(SRC_DIR, 'app'),
      APP_JS_DIR = path.join(APP_DIR, 'js'),
      APP_CSS_DIR = path.join(APP_DIR, 'css'),
      DEST_DIR = path.join(BUILD_DIR, NITROUS_ENV),
      ASSETS_DIR = path.join(DEST_DIR, 'assets');

var jsConfig = {
  name: 'js',
  entry: {
    bundle: path.join(APP_JS_DIR, 'main.coffee'),
    terminal: path.join(APP_JS_DIR, 'terminal', 'terminal.coffee')
  },
  output: {
    path: ASSETS_DIR,
    filename: '[name].js'
  },
  module: {
    loaders: [
      { test: /\.js$/, loader: 'babel', exclude: [path.join(APP_JS_DIR, 'vendor')] },
      { test: /\.coffee$/, loader: 'coffee' },
      { test: /\.emblem$/, loader: 'legacy-ember-emblem-loader', query: { appPath: APP_JS_DIR } }
    ]
  },
  resolve: {
    extensions: ['', '.js', '.coffee', '.emblem']
  },
  devtool: 'source-map',
  node: {
    fs: 'empty'
  }
};

if (NITROUS_ENV !== 'development') {
  jsConfig.plugins.push(new webpack.optimize.UglifyJsPlugin({compress:{warnings: false}}));
}

var cssConfig = {
  name: 'css',
  entry: {
    app: path.join(APP_CSS_DIR, 'application.sass'),
    terminal: path.join(APP_CSS_DIR, 'terminal.sass')
  },
  output: {
    path: ASSETS_DIR,
    filename: '[name].css'
  },
  module: {
    loaders: [
      { test: /\.s(a|c)ss$/, loader: ExtractTextPlugin.extract('style', 'css!sass') },
      { test: /\.(woff|woff2|eot|ttf|svg)(\?.*$|$)/, loader: 'file' }
    ]
  },
  resolve: {
    extensions: ['', 'css', 'sass']
  },
  plugins: [
    new ExtractTextPlugin('[name].css', {
      allChunks: true
    }),
    new OptimizeCssAssetsPlugin()
  ],
  devtool: 'source-map',
  node: {
    fs: 'empty'
  }
};

module.exports = [jsConfig, cssConfig];
