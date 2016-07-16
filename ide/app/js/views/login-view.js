App.LoginView = Em.View.extend({
  templateName: 'login',

  didInsertElement() {
    this._super();
    Em.run.scheduleOnce('afterRender', () => {
      if (window.webapp) {
        window.webapp.loading.stop();
      }
    });
  }
});
