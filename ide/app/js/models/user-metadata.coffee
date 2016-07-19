App.UserMetadata = klass = Em.Object.extend
  root_path: null
  home_path: null
  readme_url: null
  readme_opened: null
  openTabs: null

  #FIXME
  setMetadata: (data)->
    console.log data
    new Em.RSVP.Promise (resolve, reject) =>
      resolve(data)
      if data == "fail"
        reject(data)

klass.reopenClass
  fetch: ->
    new Ember.RSVP.Promise (resolve, reject) =>
      resolve(App.UserMetadata.create(
        FakeJSON.user_metadata
      ))
