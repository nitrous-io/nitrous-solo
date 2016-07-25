keysToSubmit = [
  'login'
  'user'
  'auto_close_brackets'
  'font_size'
  'highlight_line'
  'keybindings'
  'match_brackets'
  'remove_trailing_whitespace_on_save'
  'ruler_column'
  'show_line_numbers'
  'show_trailing_space'
  'soft_tabs'
  'tab_size'
  'theme'
  'wrap_lines'
  'root_path'
  'home_path'
  'readme_url'
  'readme_opened'
  'open_tabs'
]

App.User = klass = Em.Object.extend
  login: null
  user: null
  auto_close_bracket: null
  font_size: null
  highlight_line: null
  keybindings: null
  match_brackets: null
  remove_trailing_whitespace_on_save: null
  ruler_column: null
  show_line_numbers: null
  show_trailing_space: null
  soft_tabs: null
  tab_size: null
  theme: null
  wrap_lines: null
  root_path: null
  home_path: null
  readme_url: null
  readme_opened: null
  open_tabs: null

  requestSave: (delay = 5000)->
    return if @saveTimer

    @saveRequested = true
    @saveTimer = window.setTimeout(=>
      @saveTimer = null
      @save()
      @saveRequested = false
    , delay)

  save: ->
    return unless @saveRequested
    Em.$.ajax(
      type: 'PUT',
      url: '/user',
      contentType : "application/json",
      headers: { Authorization: "Token #{App.get 'token'}" },
      data: JSON.stringify(@getProperties(keysToSubmit))
    ).done((data, status, jqxhr) =>
      @setProperties(data.user)
    ).fail (jqxhr) =>
      @requestSave(1000)
