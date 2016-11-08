App.ErrorRoute = Em.Route.extend({
  beforeModel(transition) {
    if (!App.get('initError')) {
      transition.abort();
      window.location.href = '/#/ide';
    }
  }
});
