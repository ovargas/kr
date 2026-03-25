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

  // Search toggle buttons
  var toggles = document.querySelectorAll('.search-toggle');
  toggles.forEach(function(btn) {
    var param = btn.getAttribute('data-param');
    var hidden = document.querySelector('input[type="hidden"][name="' + param + '"]');

    // Sync active class from hidden input value on load
    if (hidden && hidden.value === '1') {
      btn.classList.add('active');
    }

    btn.addEventListener('click', function() {
      btn.classList.toggle('active');
      if (hidden) {
        hidden.value = btn.classList.contains('active') ? '1' : '0';
      }
    });
  });
})();
