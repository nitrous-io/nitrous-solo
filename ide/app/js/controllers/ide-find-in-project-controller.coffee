App.IdeFindInProjectController = Em.Controller.extend
  needs: ['ide', 'ideEditor']

  socketBinding: Em.Binding.oneWay 'controllers.ide.socket'
  findInProjectDialog: null
  termBinding: Ember.Binding.oneWay('findInProjectDialog.term')
  locationBinding: Ember.Binding.oneWay('findInProjectDialog.location')
  caseSensitiveBinding: Ember.Binding.oneWay('findInProjectDialog.caseSensitive')
  useRegexpBinding: Ember.Binding.oneWay('findInProjectDialog.useRegexp')
  filePatternBinding: Ember.Binding.oneWay('findInProjectDialog.filePattern')

  invalidField: null

  reset: ->
    @set 'findInProjectDialog.hide', true
    @set 'findInProjectDialog', null

  valid: ->
    @set 'invalidField', null
    if (@get('term') || '').trim() == ''
      @set 'invalidField', 'term'
      return false
    true

  showFindDialog: (location) ->
    @set 'findInProjectDialog', App.FindInProjectDialog.create location: location

  actions:
    find: ->
      if !@valid()
        App.postError "Invalid #{@get 'invalidField'}"
        return

      term = @get('term')
      location = @get('location')
      ignoreCase = !@get('caseSensitive')
      useRegexp = @get('useRegexp')
      filePattern = @get('filePattern')

      socket = @get('socket')
      editorController = @get('controllers.ideEditor')
      tab = editorController.createTab({ term: term, path: location, multiple: true }, App.SearchResults)

      cb = (res) =>
        this.reset()
        if (res.error)
          tab.addError(err)
        else if res.results?.length
          tab.addSearchResults(res.results)
        tab.doComplete()

      socket.search term, location, ignoreCase, useRegexp, filePattern, cb
