window.EmberENV =
  FEATURES:
    'query-params-new': true

Em = Ember
window.NitrousIO = Em.Application.extend Em.Evented,
  rootElement: '#app'

  platform:
    if navigator.platform.indexOf('Linux') != -1
      'linux'
    else if navigator.platform.indexOf('Windows') != -1
      'windows'
    else 'mac'

  chromeApp: false

  online: if !Em.isNone(window.navigator.onLine) then window.navigator.onLine else true
  offline: Em.computed.not 'online'

  flashMessage: null

  login: null
  systemUser: null
  homeDir: null
  settings: null

  socket: null
  pinger: null

  terminalJsPath: null
  terminalCssPath: null

  consoleFocused: false
  consoleFullscreen: false
  consoleMinimized: false
  fileTreeHidden: false

  redirecting: false

  didChangeTitle: (->
    title = @get 'title'
    document.title = if title
      "#{title} · Nitrous"
    else
      "Nitrous"
  ).observes 'title'

  didChangeSlug: (->
    slug = App.get 'slug'
    App.set 'title', "#{slug} · IDE"
  ).observes 'slug'

  saveSettings: ->
    return if @_saveSettingsTimeout
    @_saveSettingsTimeout = window.setTimeout(=>
      $.ajax
        type: 'PUT'
        url: '/user/ide_settings'
        contentType: 'application/json'
        data: JSON.stringify(@get('settings'))
        headers:
          Authorization: "Token #{@get('token')}"
      @_saveSettingsTimeout = null
    , 1000)

window.App = NitrousIO.create()
window.Lib = Ember.Namespace.create()

App.on 'load', ->
  App.set 'snackbar', App.Snackbar.create()

Em.TextSupport.reopen
  attributeBindings: ['readonly']

# HACK: silence bindAttr deprecation warning (bindAttrs are rendered by Hamlbars).
# Goes away when we can upgrade Hamlbars to 2.1+ or switch to HTMLbars/Emblem.
Em.Handlebars.registerHelper('bindAttr', Ember.Handlebars.helpers['bind-attr'])

# prevent file drops on the document causing browser to navigate away
$(document).bind 'drop dragover', (e) ->
  e.preventDefault()

# prevent esc from disconnecting websocket connection in firefox
$(window).keydown (e) ->
  e.preventDefault() if e.keyCode == 27

$(window).on 'offline', ->
  window.App?.set? 'online', false

$(window).on 'online', ->
  window.App?.set? 'online', true

$(window).on 'storage', (e) ->
  origEvent = e.originalEvent
  if origEvent.key == "token" && origEvent.newValue == null
    window.location.href = "/app/#/logout"
