var websock;

window.onload = function () {
    $('#exampleModalCenter').modal('show')
}

$('#autorization').on('click', function () {
    var username = $('#username').val()

    var xmlHttpRequest = new XMLHttpRequest();

    xmlHttpRequest.open('POST', 'checkUserName', false);
    xmlHttpRequest.send(username);

    if (xmlHttpRequest.status != 200) {
        $('#autorizationMessage').removeClass('d-none')
        return
    }

    websock = new WebSocket("ws://" + window.location.host + "/websocket?username=" + username)

    websock.onopen = function () {
        $('#exampleModalCenter').modal('hide') //hide autorize window
    }

    websock.onclose = function (event) {

        if (event.code == 4018) {
            $('#exampleModalCenter').modal('show')
        } else if (!event.wasClean) {
            alert('Код: ' + event.code + ' причина: ' + event.reason)
        }
    };

    websock.onerror = function (error) {
        alert("Ошибка " + error.readyState);
    };

    websock.onmessage = function (event) {
        var message = JSON.parse(event.data)

        var inputMessage = "<div class=\"d-flex justify-content-start mb-4\"\>\
        <div class=\"userMsg\">" +
            message.userName + ": " +
            "</div>\
        <div class=\"msgContainer\">" +
            message.messageBody +
            "\</div\>\
    </div>"
        $('#chatBox').append(inputMessage).scrollTop($('#chatBox').prop('scrollHeight'));
        log = $('#chatBox')

    }
})

function messagesHandler() {
    if (this.readyState != 4)
        return

    if (this.status != 200) {
        alert("Сообщение не отправлено")
    } else {
        var chatField = document.getElementById("chatMessages")
        chatField.value += this.responseText
    }
}

function sendMessage() {
    var xmlHttpRequest = new XMLHttpRequest();

    xmlHttpRequest.open('POST', 'sendMessage', true);

    var message = document.getElementById("messageWriter").value
    xmlHttpRequest.send(message);

    xmlHttpRequest.onreadystatechange = messagesHandler

}

$('#messageWriter').keypress(function (event) {
    if (event.which == 13) {
        $('#sendMessage').click()
        event.preventDefault();
        return true
    }
})

$('#sendMessage').on('click', function () {
    var message = $('#messageWriter').val()
    $('#messageWriter').val("")

    var inputMessage = "<div class=\"d-flex justify-content-end mb-4\"\>\
        <div class=\"msgContainer\">" +
        message +
        "\</div\>\
    </div>"
    $('#chatBox').append(inputMessage)
    scrollDown($('#chatBox'))
    var output = {}
    output.messageBody = message
    websock.send(JSON.stringify(output))

    $('#messageWriter').focus()
})

function scrollDown(element) {
    element.scrollTop($('#chatBox').prop('scrollHeight'));
}