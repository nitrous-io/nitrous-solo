window.webapp ?= {}

webapp.loading =
  start: ->
    if el = document.getElementById('webapp-loading')
      el.className = 'loading'

  stop: ->
    if el = document.getElementById('webapp-loading')
      el.className = 'loaded'
      window.setTimeout(->
        el.parentElement.removeChild(el)
      , 600)

window.setTimeout(->
  webapp.loading.start()
 , 1)
