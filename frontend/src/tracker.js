'use strict';

window.rightana = function() {
  var sessionStorageKey = 'rightana-session-key',
  trackerUrl = '',
  collectionId = '',
  debug = false,

  setup = function(trackerUrl_, collectionId_, debug_) {
    trackerUrl = trackerUrl_;
    collectionId = collectionId_;
    debug = debug_ || false;
  },

  trackPageview = function() {
    if( navigator.DonotTrack == 1 ) {
      return;
    }
    getSessionKey(sendPageView);
  },

  getSessionKey = function(cb) {
    var sessionKey = sessionStorage[sessionStorageKey];
    if (sessionKey === undefined) {
      getSessionKeyFromServer(cb);
    } else {
      cb();
    }
  },

  getSessionKeyFromServer = function(cb) {
    var d = {
      c: collectionId,
      h: location.hostname,
      bl: navigator.language,
      sr: screen.width + 'x' + screen.height,
      wr: window.innerWidth + 'x' + window.innerHeight,
      dt: getDeviceType(),
      r: document.referrer,
    }
    postDataTo(d, '/sessions', true, function(response) {
      var key = JSON.parse(response);
      sessionStorage[sessionStorageKey] = key;
      if (debug) {
        console.log('session key:', key)
      }
      cb();
    });
  },

  sendPageView = function() {
    var sessionKey = sessionStorage[sessionStorageKey];
    if (!sessionKey) {
      return;
    }
    // get the path or canonical
    var path = location.pathname + location.search;
    var canonical = document.querySelector('link[rel="canonical"]');
    if (canonical && canonical.href) {
      path = canonical.href.substring(canonical.href.indexOf('/', 7)) || '/';
    }

    var d = {
      c: collectionId,
      s: sessionKey,
      p: path,
    };
    postDataTo(d, '/pageviews', true, function() {
      if (debug) {
        console.log('post to pageviews success', path);
      }
    });
  },

  updateSessionEnd = function() {
    var sessionKey = sessionStorage[sessionStorageKey];
    if (!sessionKey) {
      return;
    }
    var d = {
      c: collectionId,
      s: sessionKey,
    };
    postDataTo(d, '/sessions/update', false, function() {
      if (debug) {
        console.log('session updated');
      }
    });
  },

  postDataTo = function(data, url, async, cb) {
    var httpRequest = new XMLHttpRequest();

    if (!httpRequest) {
      console.log('Giving up :( Cannot create an XMLHTTP instance');
      return false;
    }
    httpRequest.onreadystatechange = function() {
      if (httpRequest.readyState === XMLHttpRequest.DONE) {
        if (httpRequest.status === 200) {
          if (cb) {
            cb(httpRequest.responseText);
          }
        } else {
          console.log('failed to post ', data, ' to', url, httpRequest);
        }
      }
    }
    httpRequest.open('POST', trackerUrl + url, async);
    httpRequest.setRequestHeader('content-type', 'application/json');
    httpRequest.send(JSON.stringify(data));
  },

  getDeviceType = function() {
    var ua = navigator.userAgent,
      tablet = /Tablet|iPad/i.test(ua),
      mobile = typeof orientation !== 'undefined' || /mobile/i.test(ua);
    return tablet ? 'tablet' : mobile ? 'mobile' : 'desktop';
  },

  commands = {
    'trackPageview': trackPageview,
    'setup': setup,
  },

  processCommands = function() {
    var args = [].slice.call(arguments);
    var c = args.shift();
    commands[c].apply(this, args);
  };

  window.addEventListener('beforeunload', function(event) {
    updateSessionEnd();
  });

  (window.rightana && window.rightana.q || []).forEach(function(i) {
    processCommands.apply(this, i);
  });

  return processCommands;
}();

// for compatibility
window.k20a = function() {
  (window.k20a && window.k20a.q || []).forEach(function(i) {
    window.rightana.apply(this, i);
  });
  return window.rightana;
}();
