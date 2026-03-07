(function() {
  var source = new EventSource('/events');

  source.addEventListener('change', function() {
    location.reload();
  });

  source.addEventListener('error', function() {
    console.log('SSE connection lost, reconnecting...');
  });

  source.addEventListener('connected', function() {
    console.log('Live reload connected');
  });
})();
