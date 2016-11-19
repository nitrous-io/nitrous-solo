App.IdeFileFinderController = Em.Controller.extend
  needs: ['ide', 'ideEditor', 'ideSidebar']

  fileFinderDialog: null
  searchID: 0

  initialize: ->

  reset: ->
    @set 'fileFinderDialog', null

  showFindDialog: (path) ->
    if !path?
      path = this.get('controllers.ideSidebar.rootPath')?.absolutePath
      if !path?
        path = App.get('settings.homePath')
    @set 'fileFinderDialog', App.FileFinderDialog.create basePath: path

  startFind: (pattern, path) ->
    dialog = @get 'fileFinderDialog'
    socket = App.get('socket')
    return if !dialog || !socket

    path = App.get('settings.homePath') unless path

    @set 'findPattern', pattern
    @set 'findPath', path
    Ember.run.debounce this, this.socketFind, 500 # at most once per 0.5s

  socketFind: ->
    searchID = @incrementProperty('searchID')
    socket = App.get('socket')
    dialog = @get 'fileFinderDialog'
    pattern = @get 'findPattern'
    path = @get 'findPath'
    if dialog && socket && pattern && path
      dialog.set('status', 'searching')
      socket.fileFind pattern, path, (err, list) =>
        if @get('searchID') != searchID
          return

        if err
          dialog.setProperties(list: [], status: 'error', error: err)
          return

        dialog.setProperties(list: list.map((o) -> Em.Object.create(o)), status: 'done')

  closeFindDialog: ->
    this.reset()
