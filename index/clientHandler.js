var websock;

window.onload = function () {
    let xmlHttpRequest = new XMLHttpRequest();
    xmlHttpRequest.open('POST', 'whoami', false);
    xmlHttpRequest.send();

    if (xmlHttpRequest.status == 404) {
        $('#exampleModalCenter').modal('show')
    } else {
        let data = JSON.parse(xmlHttpRequest.response);
        let username = data.Name;
        let role = data.Role;

        connect(username, role)
    }
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

    connect(username, role)
})

function connect(username, role) {
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

        if(message.Command === 'voteStart')
        {
            voteStart();
        }

        if(message.Command === 'votedUser') {
            userVoted(message)
        }
    }
}

function userVoted(message){
    let elem = $('#user_' + snake_case_string(message.UserName) + ' .check');
    elem.show()
}

function voteStart(){
    $("#startVote").hide();
    $("#cardBlock").show();
    $(".check").hide();

    let tableRef = document.getElementById('userVotesList');
    for(let i = 0; i < tableRef.rows.length;){
        tableRef.deleteRow(i);
    }
}

function addUser(user){
    let elem = $('#user_' + snake_case_string(user.UserName) + ' .circle');
    if (elem.length) {
        elem.removeClass('darkred');
        elem.addClass('green');
        return;
    }

  let onlineCircleClass = 'darkred';

    if (user.Online === true) {
      onlineCircleClass = 'green';
  }

  let element = '\
        <span class="avatar">\
            <img class="avatar rounded-circle" src="48-512.png" alt="">\
            <span class="fa fa-circle circle ' + onlineCircleClass + '"></span>\
        </span>\
        <span>'+user.UserName+'</span><span class="fa fa-check check green" style="display: none;"></span>\
        <br>\
        <span class="user-subhead" class="role">'+user.Role+'</span>'

let tableRef = document.getElementById('userList');
let newRow = tableRef.insertRow(-1);
newRow.id = 'user_' + snake_case_string(user.UserName)
let UserCell = newRow.insertCell(0);
UserCell.innerHTML = element
}

function removeUser(user){
    var elem = $('#user_'+snake_case_string(user.UserName)+' .circle');
    elem.removeClass('green');
    elem.addClass('darkred');
}

function showResults(message) {
    $('#voteResult').html('Итогоговая оценка в часах: ' + message.voteResult);
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

    
    $("#cardBlock").hide();
    $("#startVote").show();
    $('#voteResultBlock').show()
    $('.voteCard').removeClass('active')
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

function snake_case_string(str) {
    return str && str.match(
        /[A-Z]{2,}(?=[A-Z][a-z]+[0-9]*|\b)|[A-Z]?[a-z]+[0-9]*|[A-Z]|[0-9]+/g)
        .map(s => s.toLowerCase())
        .join('_');
}