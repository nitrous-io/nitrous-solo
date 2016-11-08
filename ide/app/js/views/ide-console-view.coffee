App.IdeConsoleView = Em.View.extend App.IdeTabViewMixin,
  templateName: 'console'
  elementId: 'ide-console-view'
  classNameBindings: ['controller.fullscreen:fullscreen', 'controller.minimized:minimized']

  didInsertElement: ->
    @_super()
    @get('controller')?.restoreState()

  actions:
    didClickMinimizeConsole: ->
      if mainMenuController = App.get 'router.ideMainMenuController'
        mainMenuController.windowToggleMinimizeConsole()

    didClickFullScreen: ->
      if mainMenuController = App.get 'router.ideMainMenuController'
        mainMenuController.windowToggleFullScreenConsole()
