App.LoginController = Em.ObjectController.extend({
  login: null,
  password: null,
  errorMessage: null,

  submitDisabled: (function() {
    return !(this.get('login') && this.get('password'));
  }).property('login', 'password'),

  actions: {
    submit() {
      const login = this.get('login'),
            password = this.get('password');

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
        this.set('errorMessage', errMsg.capitalize());
      })
    }
  }
})
