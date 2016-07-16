App.LoginRoute = Em.Route.extend({
  beforeModel(transition) {
    if (App.get('token') || App.localStorage.get('token')) {
      transition.abort();
      window.location.href = '/#/ide';
    }
  }
});
