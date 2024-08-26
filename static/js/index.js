(() => {
  // node_modules/htmx.org/dist/ext/ws.js
  (function() {
    if (htmx.version && !htmx.version.startsWith("1.")) {
      console.warn("WARNING: You are using an htmx 1 extension with htmx " + htmx.version + ".  It is recommended that you move to the version of this extension found on https://extensions.htmx.org");
    }
    var api;
    htmx.defineExtension("ws", {
      /**
       * init is called once, when this extension is first registered.
       * @param {import("../htmx").HtmxInternalApi} apiRef
       */
      init: function(apiRef) {
        api = apiRef;
        if (!htmx.createWebSocket) {
          htmx.createWebSocket = createWebSocket;
        }
        if (!htmx.config.wsReconnectDelay) {
          htmx.config.wsReconnectDelay = "full-jitter";
        }
      },
      /**
       * onEvent handles all events passed to this extension.
       *
       * @param {string} name
       * @param {Event} evt
       */
      onEvent: function(name, evt) {
        var parent = evt.target || evt.detail.elt;
        switch (name) {
          // Try to close the socket when elements are removed
          case "htmx:beforeCleanupElement":
            var internalData = api.getInternalData(parent);
            if (internalData.webSocket) {
              internalData.webSocket.close();
            }
            return;
          // Try to create websockets when elements are processed
          case "htmx:beforeProcessNode":
            forEach(queryAttributeOnThisOrChildren(parent, "ws-connect"), function(child) {
              ensureWebSocket(child);
            });
            forEach(queryAttributeOnThisOrChildren(parent, "ws-send"), function(child) {
              ensureWebSocketSend(child);
            });
        }
      }
    });
    function splitOnWhitespace(trigger) {
      return trigger.trim().split(/\s+/);
    }
    function getLegacyWebsocketURL(elt) {
      var legacySSEValue = api.getAttributeValue(elt, "hx-ws");
      if (legacySSEValue) {
        var values = splitOnWhitespace(legacySSEValue);
        for (var i = 0; i < values.length; i++) {
          var value = values[i].split(/:(.+)/);
          if (value[0] === "connect") {
            return value[1];
          }
        }
      }
    }
    function ensureWebSocket(socketElt) {
      if (!api.bodyContains(socketElt)) {
        return;
      }
      var wssSource = api.getAttributeValue(socketElt, "ws-connect");
      if (wssSource == null || wssSource === "") {
        var legacySource = getLegacyWebsocketURL(socketElt);
        if (legacySource == null) {
          return;
        } else {
          wssSource = legacySource;
        }
      }
      if (wssSource.indexOf("/") === 0) {
        var base_part = location.hostname + (location.port ? ":" + location.port : "");
        if (location.protocol === "https:") {
          wssSource = "wss://" + base_part + wssSource;
        } else if (location.protocol === "http:") {
          wssSource = "ws://" + base_part + wssSource;
        }
      }
      var socketWrapper = createWebsocketWrapper(socketElt, function() {
        return htmx.createWebSocket(wssSource);
      });
      socketWrapper.addEventListener("message", function(event) {
        if (maybeCloseWebSocketSource(socketElt)) {
          return;
        }
        var response = event.data;
        if (!api.triggerEvent(socketElt, "htmx:wsBeforeMessage", {
          message: response,
          socketWrapper: socketWrapper.publicInterface
        })) {
          return;
        }
        api.withExtensions(socketElt, function(extension) {
          response = extension.transformResponse(response, null, socketElt);
        });
        var settleInfo = api.makeSettleInfo(socketElt);
        var fragment = api.makeFragment(response);
        if (fragment.children.length) {
          var children = Array.from(fragment.children);
          for (var i = 0; i < children.length; i++) {
            api.oobSwap(api.getAttributeValue(children[i], "hx-swap-oob") || "true", children[i], settleInfo);
          }
        }
        api.settleImmediately(settleInfo.tasks);
        api.triggerEvent(socketElt, "htmx:wsAfterMessage", { message: response, socketWrapper: socketWrapper.publicInterface });
      });
      api.getInternalData(socketElt).webSocket = socketWrapper;
    }
    function createWebsocketWrapper(socketElt, socketFunc) {
      var wrapper = {
        socket: null,
        messageQueue: [],
        retryCount: 0,
        /** @type {Object<string, Function[]>} */
        events: {},
        addEventListener: function(event, handler) {
          if (this.socket) {
            this.socket.addEventListener(event, handler);
          }
          if (!this.events[event]) {
            this.events[event] = [];
          }
          this.events[event].push(handler);
        },
        sendImmediately: function(message, sendElt) {
          if (!this.socket) {
            api.triggerErrorEvent();
          }
          if (!sendElt || api.triggerEvent(sendElt, "htmx:wsBeforeSend", {
            message,
            socketWrapper: this.publicInterface
          })) {
            this.socket.send(message);
            sendElt && api.triggerEvent(sendElt, "htmx:wsAfterSend", {
              message,
              socketWrapper: this.publicInterface
            });
          }
        },
        send: function(message, sendElt) {
          if (this.socket.readyState !== this.socket.OPEN) {
            this.messageQueue.push({ message, sendElt });
          } else {
            this.sendImmediately(message, sendElt);
          }
        },
        handleQueuedMessages: function() {
          while (this.messageQueue.length > 0) {
            var queuedItem = this.messageQueue[0];
            if (this.socket.readyState === this.socket.OPEN) {
              this.sendImmediately(queuedItem.message, queuedItem.sendElt);
              this.messageQueue.shift();
            } else {
              break;
            }
          }
        },
        init: function() {
          if (this.socket && this.socket.readyState === this.socket.OPEN) {
            this.socket.close();
          }
          var socket = socketFunc();
          api.triggerEvent(socketElt, "htmx:wsConnecting", { event: { type: "connecting" } });
          this.socket = socket;
          socket.onopen = function(e) {
            wrapper.retryCount = 0;
            api.triggerEvent(socketElt, "htmx:wsOpen", { event: e, socketWrapper: wrapper.publicInterface });
            wrapper.handleQueuedMessages();
          };
          socket.onclose = function(e) {
            if (!maybeCloseWebSocketSource(socketElt) && [1006, 1012, 1013].indexOf(e.code) >= 0) {
              var delay = getWebSocketReconnectDelay(wrapper.retryCount);
              setTimeout(function() {
                wrapper.retryCount += 1;
                wrapper.init();
              }, delay);
            }
            api.triggerEvent(socketElt, "htmx:wsClose", { event: e, socketWrapper: wrapper.publicInterface });
          };
          socket.onerror = function(e) {
            api.triggerErrorEvent(socketElt, "htmx:wsError", { error: e, socketWrapper: wrapper });
            maybeCloseWebSocketSource(socketElt);
          };
          var events = this.events;
          Object.keys(events).forEach(function(k) {
            events[k].forEach(function(e) {
              socket.addEventListener(k, e);
            });
          });
        },
        close: function() {
          this.socket.close();
        }
      };
      wrapper.init();
      wrapper.publicInterface = {
        send: wrapper.send.bind(wrapper),
        sendImmediately: wrapper.sendImmediately.bind(wrapper),
        queue: wrapper.messageQueue
      };
      return wrapper;
    }
    function ensureWebSocketSend(elt) {
      var legacyAttribute = api.getAttributeValue(elt, "hx-ws");
      if (legacyAttribute && legacyAttribute !== "send") {
        return;
      }
      var webSocketParent = api.getClosestMatch(elt, hasWebSocket);
      processWebSocketSend(webSocketParent, elt);
    }
    function hasWebSocket(node) {
      return api.getInternalData(node).webSocket != null;
    }
    function processWebSocketSend(socketElt, sendElt) {
      var nodeData = api.getInternalData(sendElt);
      var triggerSpecs = api.getTriggerSpecs(sendElt);
      triggerSpecs.forEach(function(ts) {
        api.addTriggerHandler(sendElt, ts, nodeData, function(elt, evt) {
          if (maybeCloseWebSocketSource(socketElt)) {
            return;
          }
          var socketWrapper = api.getInternalData(socketElt).webSocket;
          var headers = api.getHeaders(sendElt, api.getTarget(sendElt));
          var results = api.getInputValues(sendElt, "post");
          var errors = results.errors;
          var rawParameters = results.values;
          var expressionVars = api.getExpressionVars(sendElt);
          var allParameters = api.mergeObjects(rawParameters, expressionVars);
          var filteredParameters = api.filterValues(allParameters, sendElt);
          var sendConfig = {
            parameters: filteredParameters,
            unfilteredParameters: allParameters,
            headers,
            errors,
            triggeringEvent: evt,
            messageBody: void 0,
            socketWrapper: socketWrapper.publicInterface
          };
          if (!api.triggerEvent(elt, "htmx:wsConfigSend", sendConfig)) {
            return;
          }
          if (errors && errors.length > 0) {
            api.triggerEvent(elt, "htmx:validation:halted", errors);
            return;
          }
          var body = sendConfig.messageBody;
          if (body === void 0) {
            var toSend = Object.assign({}, sendConfig.parameters);
            if (sendConfig.headers)
              toSend["HEADERS"] = headers;
            body = JSON.stringify(toSend);
          }
          socketWrapper.send(body, elt);
          if (evt && api.shouldCancel(evt, elt)) {
            evt.preventDefault();
          }
        });
      });
    }
    function getWebSocketReconnectDelay(retryCount) {
      var delay = htmx.config.wsReconnectDelay;
      if (typeof delay === "function") {
        return delay(retryCount);
      }
      if (delay === "full-jitter") {
        var exp = Math.min(retryCount, 6);
        var maxDelay = 1e3 * Math.pow(2, exp);
        return maxDelay * Math.random();
      }
      logError('htmx.config.wsReconnectDelay must either be a function or the string "full-jitter"');
    }
    function maybeCloseWebSocketSource(elt) {
      if (!api.bodyContains(elt)) {
        api.getInternalData(elt).webSocket.close();
        return true;
      }
      return false;
    }
    function createWebSocket(url) {
      var sock = new WebSocket(url, []);
      sock.binaryType = htmx.config.wsBinaryType;
      return sock;
    }
    function queryAttributeOnThisOrChildren(elt, attributeName) {
      var result = [];
      if (api.hasAttribute(elt, attributeName) || api.hasAttribute(elt, "hx-ws")) {
        result.push(elt);
      }
      elt.querySelectorAll("[" + attributeName + "], [data-" + attributeName + "], [data-hx-ws], [hx-ws]").forEach(function(node) {
        result.push(node);
      });
      return result;
    }
    function forEach(arr, func) {
      if (arr) {
        for (var i = 0; i < arr.length; i++) {
          func(arr[i]);
        }
      }
    }
  })();

  // assets/js/index.js
  document.addEventListener("htmx:wsAfterMessage", function(evt) {
    var parser = new DOMParser();
    var doc = parser.parseFromString(evt.detail.message, "text/html");
    var eventEl = doc.querySelector(".event");
    var eventType = eventEl.getAttribute("data-event-type");
    if (eventType == "message") {
      var encodedMessage = eventEl.getAttribute("data-encoded-message");
      var username = eventEl.getAttribute("data-username");
      var decodedMessage = decodeMessage(encodedMessage);
      var actualChat = document.querySelector(".chat:last-child");
      var previousActualChat = document.querySelector(".chat:nth-last-child(2)");
      if (previousActualChat && username == previousActualChat.getAttribute("data-username")) {
        actualChat.remove();
        previousActualChatBubble = previousActualChat.querySelector(".chat-bubble");
        previousActualChatBubble.innerHTML += "<br />";
        previousActualChatBubble.innerHTML += decodedMessage;
      } else if (actualChat) {
        actualChatBubble = actualChat.querySelector(".chat-bubble");
        actualChatBubble.textContent = decodedMessage;
      }
    }
    let messageContainer = document.getElementById("events");
    if (messageContainer.scrollHeight - messageContainer.scrollTop < 1e3) {
      messageContainer.scrollTop = messageContainer.scrollHeight;
    }
  });
  function decodeMessage(encodedMessage) {
    encodedMessage = encodedMessage.replace(/-/g, "+").replace(/_/g, "/");
    while (encodedMessage.length % 4) {
      encodedMessage += "=";
    }
    return decodeURIComponent(atob(encodedMessage));
  }
})();
