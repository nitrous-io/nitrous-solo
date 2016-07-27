requestTempToken = require('../lib/request-temp-token.js').default
watchManager = require('../lib/watchmanager').default
Path = require('path')

App.IdeSidebarController = Em.ArrayController.extend
  needs: ['ide', 'ideFindInProject', 'ideEditor']

  homePath: null # home (~/) path
  rootPath: null # current root path in the sidebar
  content:  null # array of all path objects
  uploadDialog: null
  confirmDeletePathDialog: null

  socketBinding: Em.Binding.oneWay 'App.router.ideController.socket'
  socketReadyBinding: Em.Binding.oneWay 'socket.started'
  socketNotReady: Em.computed.not 'socketReady'

  activeFilesMonitorBinding: Em.Binding.oneWay 'App.router.ideController.activeFilesMonitor'

  repositoryBinding: Em.Binding.oneWay 'App.router.indexontroller.repository'

  hiddenFilesVisible: false

  currentlySelectedPath: null

  # attributes used for drag&drop
  currentlyDraggedPath: null
  currentDroppable: null

  initialize: ->
    @set 'homePath', homePath = App.Path.create absolutePath: App.get('userMetadata').home_path, directory: true
    @set 'content', Em.A([homePath])
    @set 'rootPath', App.Path.create absolutePath: App.get('userMetadata').root_path, directory: true

  reset: ->
    @set 'homePath', null
    @set 'rootPath', null
    @set 'content', null
    @set 'uploadDialog', null
    @set 'currentlyDraggedPath', null
    @set 'currentDroppable', null

  socketDidStartForTheFirstTime: (->
    if @get('socket.connectionCount') == 1
      Em.run.next =>
        socket = @get 'socket'
        @openDirectory @get('rootPath')
        @restoreOpenTabs()
  ).observes 'socket.connectionCount'

  socketDidBecomeReady: (->
    if @get('socketReady')
      # reset watchmanager and rewatch watched paths
      watchManager.reset()
      rootPath = @get('rootPath')
      if rootPath && rootPath.get('expanded')
        @watchAndFetchRecursively(rootPath)
  ).observes 'socketReady'

  restoreOpenTabs: ->
    if openTabs = App.get('userMetadata').openTabs
      index = 0
      for tab in openTabs
        if tab.path
          pth = App.Path.create absolutePath: tab.path, directory: false
          this.togglePath pth,
            error: false
            active: tab.active
            index: index
            scrollPosition: tab.scrollPosition
            selections: tab.selections
        else
          continue
        index++

  # settings options:
  #   done: callback (document)
  #   error: callback (errorMessage)
  togglePath: (path, settings={}) ->
    return if path.get('loading') || !@get('socketReady')

    if path.get 'directory'
      if path.get('expanded')
        @closeDirectory(path)
      else
        @openDirectory(path)
    else
      @openFile(path, settings)

  openFile: (path, settings={}) ->
    return if path.get('loading') || !@get('socketReady')

    if settings.error != false && typeof settings.error != 'function'
      settings.error = (err) ->
        message = switch err
          when 'file_too_big'
            'Cannot open the selected file because it is too large.'
          else
            "Cannot open the selected file - #{err.replace(/_/g, ' ')}."
        App.postError message

    editorController = @get('controllers.ideEditor')
    if doc = editorController.findProperty 'path', path
      unless settings.active == false
        editorController.makeTabActive doc
      if settings.done
        settings.done(doc)
    else
      @fetchFileContent path, settings

  openDirectory: (path) ->
    return if path.get('loading') || !@get('socketReady')

    path.set('expanded', true)
    @watchAndFetchRecursively(path)

  closeDirectory: (path) ->
    return if path.get('loading') || !@get('socketReady')

    path.set('expanded', false)
    @unwatchRecursively(path)

  downloadPath: (path) ->
    if !path.get('loading')
      requestTempToken().then (token) =>
        url = "/files?token=#{token}&path=#{path.get('absolutePath')}"
        downloadFrame = $('<iframe></iframe>')
        downloadFrame.css 'display', 'none'
        downloadFrame.attr 'src', url
        $('body').append downloadFrame
        window.setTimeout(->
          downloadFrame.remove()
        , 5000)
      , (jqxhr) ->
        App.postError(jqxhr.responseJSON, 'Could not download file, please try again.')

  reveal: (path) ->
    return unless path.get
    origPath = path
    while parent = path.get 'parentAbsolutePath'
      path = App.Path.create absolutePath: parent, directory: true
      unless path.get 'expanded'
        @togglePath path
    origPath.set 'showInSidebar', true
    if previouslySelectedPath = @get 'currentlySelectedPath'
      previouslySelectedPath.set 'selected', false
    origPath.set 'selected', true
    @set 'currentlySelectedPath', origPath

  fetchFileList: (path, cb) ->
    socket = @get 'socket'
    path.incrementProperty 'loadCount'
    socket.fetchFileList path.get('absolutePath'), (err, content) ->
      if err
        path.set 'expanded', false
      else
        path.updateContent content
        if typeof cb == 'function'
          cb()
      path.decrementProperty 'loadCount'

  # reloads document content of a tab with a given path from server
  # this is for operations where we modified content of the file on the server
  # e.g. because of running gofmt/goimports on the code, and we want those
  # changes be reflected in the editor
  refreshTextFileContent: (path, settings = {}) ->
    socket = @get 'socket'
    path.incrementProperty 'loadCount'
    editorController = @get('controllers.ideEditor')

    socket.fetchFileContent path.get('absolutePath'), (err, contentType, content) =>
      path.decrementProperty 'loadCount'
      if err
        if settings.error
          settings.error(err)
        return

      editorController = @get('controllers.ideEditor')
      if doc = editorController.findProperty 'path', path
        doc.set 'replaceContentTrigger', content
      if settings.done
        settings.done(tab)

  fetchFileContent: (path, settings = {}) ->
    socket = @get 'socket'
    path.incrementProperty 'loadCount'
    editorController = @get('controllers.ideEditor')

    socket.fetchFileContent path.get('absolutePath'), (err, contentType, content) =>
      path.decrementProperty 'loadCount'

      if err && settings.error
        settings.error(err)
        return

      tabClass = if contentType?.match /^image/ then App.Image else App.Document
      tab = editorController.createTab(
        content: content
        path: path
        contentType: contentType
        active: settings.active
        index: settings.index
        selections: settings.selections
        scrollPosition: settings.scrollPosition
      , tabClass)
      if settings.done
        settings.done(tab)

  refresh: (path) ->
    if path.get('directory') && path.get('expanded')
      delay = 0

      for subpath in path.get 'arrangedContent'
        if subpath.get('directory') && subpath.get('expanded')
          delay += @constructor.REFRESH_DELAY
          ((controller, subpath, delay) ->
            window.setTimeout ->
              controller.refresh subpath
            , delay
          )(this, subpath, delay)

      unless path.get 'loading'
        this.fetchFileList path

  refreshPath: (pathString) ->
    homePath = @get 'homePath'
    rootPath = @get 'rootPath'
    if path = (homePath.findByRelativePath(pathString) || rootPath.findByRelativePath(pathString))
      this.refresh path

    if fileMon = @get 'activeFilesMonitor'
      fileMon.refresh()

  refreshAll: ->
    this.refresh @get('rootPath')

  findInFolder: (path) ->
    controller = @get 'controllers.ideFindInProject'
    if controller
      controller.showFindDialog(path.absolutePath)

  setHomeAsRoot: ->
    @set 'rootPath', @get('homePath')
    @set 'rootPath.expanded', true
    @refresh @get 'rootPath'

  setPathAsRoot: (path) ->
    @set 'rootPath', path

  rootPathDidChange: (->
    val = @get 'rootPath'
    return if App.get('userMetadata').root_path == val
    return if App.get('userMetadata').root_path == val
    App.get('userMetadata').setMetadata({rootPath: val.get 'collabPath'})
  ).observes 'rootPath'

  newPath: (path, directory) ->
    this.togglePath path if !path.get 'expanded'
    newPath = App.Path.create
      absolutePath: "#{path.get('absolutePath')}/__aio_untitled_path__"
      new: true
      editable: true
      directory: directory
    path.addObject newPath

  saveNewPath: (path, name) ->
    if parentPath = @findParentPath(path)
      parentAbsolutePath = path.get 'parentAbsolutePath'
      socket = @get 'socket'
      absolutePath = _.chain([parentAbsolutePath, name]).compact().join('/').value()

      callback = (err) =>
        if err
          errMsg = if err.match /file_exists_at_path/
            "\"#{name}\" already exists. Please enter a different filename."
          else
            "Error occurred while trying to save the file."
          App.postError errMsg
        path.set 'new', false # temp paths with 'new: false' get removed on refresh
        this.refresh parentPath

      if path.get 'directory'
        socket.makeDirectory absolutePath, callback
      else
        socket.writeFile absolutePath, false, '', callback

    @get('controllers.ideEditor').trackActivity()

  cancelNewPath: (path) ->
    if parentPath = @findParentPath(path)
      parentPath.get('content').removeObject path

  renamePath: (path) ->
    path.set 'editable', true
    @get('controllers.ideEditor').trackActivity()

  # apply rename operations to path
  # - path: the path object that we want to rename
  # - newPath: the path object / new name of the path.
  # - isDrag: if true, newPath will be the new path object that we drag the
  # path into a.k.a new parent path
  applyRenamedPath: (path, newPath, isDrag) ->
    if typeof newPath == "string"
      return this.resetRenamedPath path if newPath == path.get 'name'
    if isDrag && !(newPath instanceof App.Path)
      return console.warning("expected a App.Path object")
    if !isDrag && !(typeof newPath == "string")
      return console.warning("expected a path string")

    rootPath = @get('rootPath')
    parentAbsolutePath = path.get 'parentAbsolutePath'
    # current directory that hold the path
    parentPath = rootPath.findByRelativePath parentAbsolutePath
    return unless parentPath

    socket = @get 'socket'
    currentPathName = path.get 'name'

    if isDrag
      # if this is a drag operation, newPath is going to be the path that we drag
      # the file into, while the name of the file will be kept
      newParentPath = newPath.get('absolutePath')
      newPathName = currentPathName
    else
      # otherwise, this is the rename operation, which the parentPath will be
      # kept and the file will have a new name
      newParentPath = parentAbsolutePath
      newPathName = newPath

    currentAbsolutePath = this._makeAbsolutePath parentAbsolutePath, currentPathName
    newAbsolutePath = this._makeAbsolutePath newParentPath, newPathName

    path.incrementProperty 'loadCount'
    socket.rename currentAbsolutePath, newAbsolutePath, (err) =>
      if err
        if err == 'aio:name_taken'
          App.set 'flashMessage', { type: 'error', text: "The name “#{newPathName}” is already taken. Please choose a different name." }
        else
          App.set 'flashMessage', { type: 'error', text: 'Rename failed.' }
      else
        pathIsDir = path.get('directory')
        @unwatchRecursively(path) if pathIsDir
        path.set 'absolutePath', newAbsolutePath
        if pathIsDir && path.get('expanded')
          @watchAndFetchRecursively(path)

      path.decrementProperty 'loadCount'
      path.set 'editable', false

      this.refresh parentPath
      this.refresh newPath if isDrag

  resetRenamedPath: (path) ->
    path.set 'editable', false

  showUploadDialog: (path) ->
    path ?= @get 'rootPath'
    @set 'uploadDialog', App.UploadDialog.create path: path

  beginUpload: ->
    uploadDialog = @get 'uploadDialog'
    if uploadDialog && uploadDialog.get('length')
      uploadDialog.set 'uploading', true
      uploadDialog.forEach (upload) ->
        upload.submit()
    else
      @set 'uploadDialog', null

  didFinishAnUpload: (path) ->
    if path
      path = App.Path.create absolutePath: path
      this.refreshPath path.get('parentAbsolutePath')
    uploadDialog = @get 'uploadDialog'
    if uploadDialog && uploadDialog.get 'completed'
      uploadDialog.set('hide', true)
      @set 'uploadDialog', null

  confirmDelete: (path) ->
    @set 'confirmDeletePathDialog', App.ConfirmDeletePathDialog.create path: path
    @get('controllers.ideEditor').trackActivity()

  delete: (path) ->
    return if path == @get 'rootPath'

    deleteMethod = if path.get('directory') then 'deleteDir' else 'deleteFile'
    socket = @get 'socket'
    parentPath = @findParentPath path
    path.set 'beingDeleted', true
    path.incrementProperty 'loadCount'

    socket[deleteMethod] path.get('absolutePath'), (err) =>
      if err
        path.decrementProperty 'loadCount'
        path.set 'beingDeleted', false
        App.postError(err)
      else
        @unwatchRecursively(path) # should not be necessary due to onStop, but just in case.
        path.set 'deleted', true
        if parentPath
          parentPath.get('content').removeObject path
        # TODO: not strictly necessary but, remove all children paths recursively from 'content' too?
        this.get('content').removeObject path

  moveDraggedPathTo: (path) ->
    draggedPath = @get 'currentlyDraggedPath'
    absoluteDroppablePath = path.get 'absolutePath'
    #expand the droppable path
    path.set 'expanded', true
    this.applyRenamedPath draggedPath, path, true
    @set 'currentlyDraggedPath', null
    @set 'currentDroppable', null

  _makeAbsolutePath: (parentAbsolutePath, fileName) ->
    _.chain([parentAbsolutePath, fileName]).compact().join('/').value()

  findParentPath: (path) ->
    rootPath = @get 'rootPath'
    parentAbsolutePath = path.get 'parentAbsolutePath'
    return rootPath.findByRelativePath parentAbsolutePath

  # recursively watch/fetch a folder, and all "expanded" sub-folders
  watchAndFetchRecursively: (pathObj) ->
    absPath = pathObj.get('absolutePath')
    @fetchFileList pathObj, =>
      watchManager.unlessWatching absPath, =>
        onStart = (id) ->
          watchManager.add(id, absPath)

        onChange = (id, evt, path) =>
          return if evt == "modified"
          watchManager.debounce id, 100, =>
            @refreshPath(Path.dirname(path))

        onStop = (id) =>
          @unwatchRecursively(watchManager.pathForID(id))

        @get('socket').watch(absPath, onStart, onChange, onStop)

      if content = pathObj.get('content')
        for subpath in content
          if subpath.get('directory') && subpath.get('expanded')
            @watchAndFetchRecursively(subpath)

  unwatchRecursively: (path) ->
    if path instanceof App.Path
      path = path.get('absolutePath')

    watchManager.ifWatching path, =>
      if ids = watchManager.idsForSubPaths(path)
        socket = @get 'socket'
        for id in ids
          socket.unwatch(id)
          watchManager.remove(id)

App.IdeSidebarController.reopenClass
  REFRESH_DELAY: 200
