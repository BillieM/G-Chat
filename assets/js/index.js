import 'htmx.org/dist/ext/ws.js';

var messageForm = document.getElementById("message-form")

// handles decoding message/ notification events & performing formatting 
document.addEventListener('htmx:wsAfterMessage', function(evt) {

    // Parse the HTML string
    var parser = new DOMParser();
    var doc = parser.parseFromString(evt.detail.message, 'text/html');

    var eventEl = doc.querySelector(".event")

    // get event type
    var eventType = eventEl.getAttribute("data-event-type");

    if (eventType == "message") {
        var encodedMessage = eventEl.getAttribute('data-encoded-message');
        var username = eventEl.getAttribute('data-username');
        var messageType = eventEl.getAttribute('data-type');
        
        // Decode the message
        var decodedMessage = decodeMessage(encodedMessage);
        
        var actualChat = document.querySelector('.chat:last-child');
        var previousActualChat = document.querySelector('.chat:nth-last-child(2)')
    
        if (previousActualChat && 
            username == previousActualChat.getAttribute("data-username") && 
            messageType == previousActualChat.getAttribute("data-type")
        ) {
            // merge to above message
            actualChat.remove()
            previousActualChatBubble = previousActualChat.querySelector(".chat-bubble")
            previousActualChatBubble.innerHTML += "<br />"
            previousActualChatBubble.innerHTML += decodedMessage
        } else if (actualChat) {
            actualChatBubble = actualChat.querySelector(".chat-bubble")
            actualChatBubble.textContent = decodedMessage;
        }
    }
    
    // Get the div element
    let eventContainer = document.getElementById('events');
    // Scroll to the bottom of the div if already at bottom
    if (eventContainer.scrollHeight - eventContainer.scrollTop < 1000) {
        eventContainer.scrollTop = eventContainer.scrollHeight;
    }
});   

document.addEventListener("htmx:wsBeforeSend", function(evt) {
    var messageText = messageForm.querySelector("#message-text")
    if (messageText.value.length == 0) {
        evt.preventDefault()
    }
})

document.addEventListener('htmx:wsAfterSend', function(evt) {
    var messageText = messageForm.querySelector("#message-text")
    messageText.value = ""
})


function decodeMessage(encodedMessage) {
    // First, replace any URL-safe base64 characters
    encodedMessage = encodedMessage.replace(/-/g, '+').replace(/_/g, '/');
    
    // Add padding if necessary
    while (encodedMessage.length % 4) {
        encodedMessage += '=';
    }
    
    // Decode the base64 string
    return decodeURIComponent(atob(encodedMessage));
}
