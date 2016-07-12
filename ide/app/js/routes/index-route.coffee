App.IndexRoute = Em.Route.extend
  beforeModel: (transition) ->
    window.location.href = '/#/ide'
