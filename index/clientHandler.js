var websock;

window.onload = function () {
    $('#exampleModalCenter').modal('show')
}

$('#autorization').on('click', function () {
    let username = $('#username').val()
    let role = $('#role').val()
    let xmlHttpRequest = new XMLHttpRequest();

    xmlHttpRequest.open('POST', 'checkUserName', false);
    xmlHttpRequest.send(username);

    if (xmlHttpRequest.status != 200) {
        $('#autorizationMessage').removeClass('d-none')
        return
    }

    websock = new WebSocket("ws://" + window.location.host + "/websocket?username=" + username + '&role=' + role)

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
        if (message.hasOwnProperty('voteResult')) {
            showResults(message);
        }

        if (message.Command === 'connect') {
            addUser(message);   
        }
        if (message.Command === 'disconnect') {
            removeUser(message);    
        }
    }
})

function addUser(user){
  let element = '<img class="avatar  rounded-circle" src="48-512.png" alt="">\
        <span>'+user.UserName+'</span>\
        <br>\
        <span class="user-subhead" class="role">'+user.Role+'</span>'

let tableRef = document.getElementById('userList');
let newRow = tableRef.insertRow(-1);
newRow.id = 'user_' + user.UserName
let UserCell = newRow.insertCell(0);
UserCell.innerHTML = element
}

function removeUser(user){
    var elem = $('#user_'+user.UserName);
    elem.remove();
}

function showResults(message) {
    $('#voteResult').html('Итогоговая оценка в днях: ' + message.voteResult);
    users = new Map(Object.entries(message.votes))
    users.forEach(element => {
        let tableRef = document.getElementById('userVotesList');
        // Insert a row in the table at row index 0
        let newRow = tableRef.insertRow(0);

        // Insert a cell in the row at index 0
        let UserNameCell = newRow.insertCell(0);
        let voteCell = newRow.insertCell(1);

        // Append a text node to the cell
        let newText = document.createTextNode(element.userName);
        UserNameCell.appendChild(newText);
        if (element.vote.isCoffeeBreak) {
            voteCell.innerHTML = '<span class="fa fa-coffee"></span>'
        } else if (element.vote.isQuestionMark) {
            voteCell.innerHTML = '<span class="fa fa-question"></span>'
        } else {
            let newText = document.createTextNode(element.vote.value);
            voteCell.appendChild(newText);
        }

    });
    $('#voteResultBlock').show()
}

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

$('#startVote').keypress(function (event) {
    if (event.which == 13) {
        $('#startVote').click()
        event.preventDefault();
        return true
    }
})

$('#startVote').on('click', function () {
    let output = {}
    output.command = 'startVote'
    let body = {}
    body.topic = $('#topicName').val()
    output.body = body
    $('#voteResultBlock').hide()
    websock.send(JSON.stringify(output))
})

$('.voteCard').on('click', function () {
    $('.voteCard').not(this).removeClass('active')

    let output = {}
    output.command = 'vote'
    output.body = {
        value: parseFloat($(this).val()),
        isCoffeeBreak: $(this).data('coffee'),
        isQuestionMark: $(this).data('question')
    }

    websock.send(JSON.stringify(output))
})
