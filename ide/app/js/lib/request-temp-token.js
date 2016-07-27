export default function requestTempToken() {
  const promise = new Em.RSVP.Promise(function (resolve, reject) {
    Em.$.ajax({
      type: 'POST',
      url: "/token/temp",
      headers: {
        Authorization: `Token ${App.localStorage.get('token')}`
      }
    }).done(function (data, status, jqxhr) {
      resolve(data.token);
    }).fail(function (jqxhr) {
      reject(jqxhr);
    })
  });
  return promise;
};
