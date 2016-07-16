App.ErrorView = Em.View.extend({
  templateName: 'error',

  didInsertElement() {
    this._super();
    Em.run.scheduleOnce('afterRender', () => {
      if (window.webapp) {
        window.webapp.loading.stop();
      }
    });
  },

  actions: {
    retry() {
      location.replace('/');
    }
  }
});
