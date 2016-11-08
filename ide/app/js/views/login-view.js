App.LoginView = Em.View.extend({
  classNames: ['login-view'],
  templateName: 'login',

  didInsertElement() {
    this._super();
    Em.run.scheduleOnce('afterRender', () => {
      if (window.webapp) {
        window.webapp.loading.stop();
      }
      $('#login-login').focus();
    });
  }
});
