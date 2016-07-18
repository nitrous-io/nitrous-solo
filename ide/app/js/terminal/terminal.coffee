require '!!script!../vendor/polyfills'
require '!!script!../vendor/term'

parentWindow = window.parent
parentWindow = null if parentWindow == window

isMac = navigator.userAgent.indexOf("Macintosh") != -1
chromeApp = parentWindow?.App?.get('chromeApp') || false

decodeUTF8 = (s) ->
  decodeURIComponent(escape(s))

window.NitrousTerminal =
  term: null
  firstConnect: true
  readonly: false
  ide: false

  resizeTimeout: null

  dataBuffer: null
  bufferTimeout: null

  init: ->
    @term = term = new Terminal(80, 25)
    term.open(document.body)
    term.sizeToFit()
    term.showCursor()
    term.convertEol = true

    term.originalKeyDown = @term.keyDown
    term.keyDown = (event) =>
      if @ide
        r = switch
          when @matchesFocusEditorKeybinding event
            @focusEditor event
          when @matchesFullscreenConsoleKeybinding event
            @toggleFullscreenConsole event
          when @matchesMinimizeConsoleKeybinding event
            @toggleMinimizeConsole event
          when @matchesNextTabKeybinding event
            @makeNextTabActive event
          when @matchesPreviousTabKeybinding event
            @makePreviousTabActive event
          when @matchesNewConsoleScreenKeybinding event
            @newTab event
          when @matchesCloseConsoleScreenKeybinding event
            @closeTab event
          when @matchesCtrlTabKeybinding event
            false
          when @matchesCtrlShiftTabKeybinding event
            false
          else
            true
        return false if !r
      term.originalKeyDown(event)

    term.on 'resize', (cols, rows) =>
      @onResize(cols, rows)

    term.on 'title', (title) =>
      @didChangeTitle(title)

    term.handler = (data) =>
      return if @readonly
      @onKeyDown(data)

    onresize = =>
      window.onresize = null
      return unless @term
      if @resizeTimeout
        window.clearTimeout(@resizeTimeout)
      @resizeTimeout = window.setTimeout(=>
        @resizeTimeout = null
        @term.sizeToFit()
      , 50)
      window.onresize = onresize
    onresize()

    window.onfocus = =>
      return unless @term
      Terminal.focus = @term
      @term.redrawCursor()
      parentWindow?.App?.set?('consoleFocused', true)

    window.onblur = =>
      return unless @term
      Terminal.focus = null
      @didBlur()
      @term.redrawCursor()
      parentWindow?.App?.set?('consoleFocused', false)

    document.addEventListener('copy', @handleCopy, false)

  destroy: ->
    @term?.destroy()

  focus: ->
    window.focus()
    @term?.focus()

  getSize: ->
    { cols: @term.cols, rows: @term.rows }

  didBlur: ->
    # override me

  didChangeTitle: (title) ->
    # override me

  clear: ->
    @term.write '\x1b[r\x1b[H\x1b[J'

  onKeyDown: ->
  onResize: ->

  handleCopy: ->
    sel = window.getSelection?()
    range = sel.getRangeAt?(0)
    text = sel.toString?()
    tempEl = document.createElement('pre')
    # return if browser required features are missing
    return true unless sel && sel.selectAllChildren && sel.removeAllRanges && sel.addRange && range && text && tempEl.textContent?
    # make temp element invisible
    tempEl.style.whiteSpace = 'pre'
    tempEl.style.position = 'fixed'
    tempEl.style.left = '100%'
    tempEl.style.opacity = '0'
    # strip trailing whitespaces
    tempEl.textContent = text.replace(/\s+$/gm, '')
    document.body.appendChild(tempEl)
    sel.selectAllChildren(tempEl)
    window.setTimeout(->
      tempEl.parentNode.removeChild(tempEl)
      sel.removeAllRanges()
      sel.addRange(range)
    , 10)
    true

  handleData: (data) ->
    @dataBuffer ?= ''
    window.clearTimeout(@bufferTimeout)
    @dataBuffer += data
    decoded = @dataBuffer
    try
      decoded = decodeUTF8(@dataBuffer)

    flush = =>
      @dataBuffer = ''
      @term.write(decoded)

    # if invalid utf8 char is detected, buffer with timeout
    if decoded.length > 0 && decoded[decoded.length - 1] == "\uFFFD"
      @bufferTimeout = window.setTimeout(=>
        flush()
        @bufferTimeout = null
      , 200)
      return

    flush()

  didOpen: () ->
    @term.sizeToFit()
    if @firstConnect
      @firstConnect = false
      @clear()

  didFail: (err) ->

  didExit: (data) ->
    @term?.write("\r\n[Session closed - #{data}]")

  focusEditor: (ev) ->
    return true unless parentWindow?.App?.get
    Terminal.cancel(ev)
    parentWindow.App.get('router.ideMainMenuController')?.windowFocusEditor()
    false

  toggleFullscreenConsole: (ev) ->
    return true unless parentWindow?.App?.get
    Terminal.cancel(ev)
    parentWindow.App.get('router.ideMainMenuController')?.windowToggleFullScreenConsole()
    false

  toggleMinimizeConsole: (ev) ->
    return true unless parentWindow?.App?.get
    Terminal.cancel(ev)
    parentWindow.App.get('router.ideMainMenuController')?.windowToggleMinimizeConsole()
    false

  makeNextTabActive: (ev) ->
    return true unless parentWindow?.App?.get
    Terminal.cancel(ev)
    parentWindow.App.get('router.ideMainMenuController')?.windowSelectNextTab()
    false

  makePreviousTabActive: (ev) ->
    return true unless parentWindow?.App?.get
    Terminal.cancel(ev)
    parentWindow.App.get('router.ideMainMenuController')?.windowSelectPreviousTab()
    false

  newTab: (ev) ->
    return true unless parentWindow?.App?.get
    Terminal.cancel(ev)
    parentWindow.App.get('router.ideMainMenuController')?.fileNew()
    false

  closeTab: (ev) ->
    return true unless parentWindow?.App?.get
    Terminal.cancel(ev)
    parentWindow.App.get('router.ideMainMenuController')?.fileClose()
    false

  matchesFocusEditorKeybinding: (event) ->
    (isMac && event.shiftKey && event.ctrlKey && event.keyCode == 67) || # mac: shift+ctrl+c
    (!isMac && event.ctrlKey && event.altKey && event.keyCode == 67)  || # pc:  ctrl+alt+c
    (chromeApp && isMac && event.metaKey && event.keyCode == 84)  || # mac chrome app: command+t
    (chromeApp && !isMac && event.shiftKey && event.ctrlKey && event.keyCode == 84) # pc chrome app: shift+ctrl+t

  matchesFullscreenConsoleKeybinding: (event) ->
    (event.keyCode == 70 && (
      (isMac && event.shiftKey && event.ctrlKey) || # mac: shift+ctrl+f
      (!isMac && event.ctrlKey && event.altKey)     # pc:  ctrl+alt+f
    ))

  matchesMinimizeConsoleKeybinding: (event) ->
    (event.keyCode == 77 && (
      (isMac && event.shiftKey && event.ctrlKey) || # mac: shift+ctrl+m
      (!isMac && event.ctrlKey && event.altKey)     # pc:  ctrl+alt+m
    ))

  matchesNextTabKeybinding: (event) ->
    (!chromeApp && event.shiftKey && event.ctrlKey && event.keyCode == 221) ||      # mac/pc: shift+ctrl+]
    (chromeApp && !event.shiftKey && event.ctrlKey && event.keyCode == 9) ||        # mac/pc chrome app: ctrl+tab
    (chromeApp && isMac && event.shiftKey && event.metaKey && event.keyCode == 221) # mac chrome app: shift+command+]

  matchesPreviousTabKeybinding: (event) ->
    (!chromeApp && event.shiftKey && event.ctrlKey && event.keyCode == 219) ||      # mac/pc: shift+ctrl+[
    (chromeApp && event.shiftKey && event.ctrlKey && event.keyCode == 9) ||         # mac/pc chrome app: shift+ctrl+tab
    (chromeApp && isMac && event.shiftKey && event.metaKey && event.keyCode == 219) # mac chrome app: shift+command+[

  matchesNewConsoleScreenKeybinding: (event) ->
    (event.keyCode == 78 && (
      (isMac && event.shiftKey && event.ctrlKey) || # mac: shift+ctrl+n
      (!isMac && event.ctrlKey && event.altKey)  || # pc:  ctrl+alt+n
      (chromeApp && isMac && event.metaKey)      || # mac chrome app: command+n
      (chromeApp && !isMac && event.shiftKey && event.ctrlKey) # pc chrome app: shift+ctrl+n
    ))

  matchesCloseConsoleScreenKeybinding: (event) ->
    (event.keyCode == 87 && (
      (isMac && event.shiftKey && event.ctrlKey) || # mac: shift+ctrl+w
      (!isMac && event.ctrlKey && event.altKey)  || # pc:  ctrl+alt+w
      (chromeApp && isMac && event.metaKey)      || # mac chrome app: command+w
      (chromeApp && !isMac && event.shiftKey && event.ctrlKey) # pc chrome app: shift+ctrl+w
    ))

  # keybinding ctrl+tab
  matchesCtrlTabKeybinding: (event) ->
    event.ctrlKey && event.keyCode == 9

  # keybinding ctrl+shift+tab
  matchesCtrlShiftTabKeybinding: (event) ->
    event.ctrlKey && event.shiftKey && event.keyCode == 9

window.webFontLoaded = ->
  NitrousTerminal.term?.sizeToFit()
