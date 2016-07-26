window.TabWatcher = class @TabWatcher
  instance = null
  constructor: (@rootPane, @user) ->
    @changedAt = null
    @lastSynced = null
    instance ?= this

  notifyOfChange: ->
    @scheduleSync()

  scheduleSync: (delay = 25000)->
    return if @timer
    @timer = window.setTimeout(=>
      @timer = null
      @syncChanges()
    , delay)

  syncChanges: ->
    allTabs = @rootPane.allTabs()
    changes = []
    for tab in allTabs
      pathValue = null
      path = tab.get('path')
      if path
        if typeof path is 'string'
          pathValue = path
        else
          pathValue = path.get('absolutePath')

        change = {
          path: pathValue,
          scroll_position: tab.get('scrollPosition'),
          selections: tab.get('selections'),
          active: tab.get('active'),
          readme: tab.get('readme')
        }
        changes.push change

    @user.open_tabs = changes
    @user.requestSave()

  @notifyOfChange: ->
    return unless instance
    instance.notifyOfChange()

  @syncChanges: ->
    return unless instance
    instance.syncChanges()
