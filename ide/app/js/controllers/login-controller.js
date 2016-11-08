App.LoginController = Em.ObjectController.extend({
  login: null,
  password: null,
  submitting: false,

  submitDisabled: (function() {
    return !(this.get('login') && this.get('password')) || this.get('submitting');
  }).property('login', 'password', 'submitting'),

  actions: {
    submit() {
      console.log('submit');
      const login = this.get('login'),
            password = this.get('password');

      this.set('submitting', true);
      $.ajax({
        type: 'POST',
        url: '/token',
        data: { login, password }
      }).then((res) => {
        App.localStorage.set('token', res.token);
        this.transitionTo('ide');
      }, (jqxhr) => {
        let errMsg;
        if (jqxhr.responseJSON) {
          errMsg = jqxhr.responseJSON.error_description || jqxhr.responseJSON.error;
        }
        errMsg = errMsg || `Unexpected response: ${jqxhr.status} ${jqxhr.statusText}`;
        App.postError(errMsg.capitalize());
        this.set('submitting', false);
        Em.run.scheduleOnce('afterRender', () => {
          $('#login-login').select().focus();
        });
      });
    }
  }
})
