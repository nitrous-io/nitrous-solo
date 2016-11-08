App.Pinger = klass = Em.Object.extend
  latency: 0
  container: null

  init: ->
    @_super.apply this, arguments

    @ping()
    window.setInterval(=>
      @ping()
    , @constructor.pingInterval)

  ping: ->
    t = +new Date
    Em.$.ajax(
      type: 'GET'
      headers: { Authorization: "Token #{App.get 'token'}" }
      url: '/ping'
    ).then =>
      t2 = +new Date
      @set 'latency', t2 - t
    , =>
      @set 'latency', -1

klass.reopenClass
  pingInterval: 10000 # millisecond
