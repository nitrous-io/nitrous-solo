App.FileFinderDialog = klass = Em.Object.extend
  pattern: null
  basePath: null

  list: null
  status: null

  selectedIndex: 0

  error: null

  humanStatus: (->
    s = switch @get 'status'
      when 'error'
        if error = @get('error') then error else 'Error finding'
      when 'done'
        if @get('list.length') == 0
          'Not found'
        else
          null
      when 'searching'
        'Searching...'
      else
        null
  ).property 'status', 'list.length'

  placeholder: (->
    "Find files in #{@get 'prettyPath'}"
  ).property 'prettyPath'

  prettyPath: (->
    path = @get 'basePath'
    if path == '.'
      'all directories'
    else
      if path.slice(-1) == '/' then path else path + '/'
  ).property 'basePath'

  incrementIndex: ->
    index = @get 'selectedIndex'
    list = (@get 'list') || []
    @incrementProperty('selectedIndex') if index + 1 < list.length

  decrementIndex: ->
    index = @get 'selectedIndex'
    list = (@get 'list') || []
    @decrementProperty('selectedIndex') if index - 1 >= 0

  selectedFile: (->
    list = @get 'list'
    return if !list
    list.objectAt(@get 'selectedIndex')
  ).property 'list.@each.filename', 'selectedIndex'

  updateSelected: (->
    list = @get 'list'
    return if !list
    selectedIndex = @get 'selectedIndex'
    list.forEach (result, index) ->
      result.set 'selected', index == selectedIndex
  ).observes 'list.@each.filename', 'selectedIndex'

  resetSelected: (->
    @set 'selectedIndex', 0
  ).observes 'list'
