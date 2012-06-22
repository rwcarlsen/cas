
function submitData(eventObj) {
  box = $("#databox")
  data = box.val()
  $.post("/cas/put", data, printResponse)
}

function getData(eventObj) {
  box = $("#refbox")
  data = box.val()
  $.post("/cas/get", data, printResponse)
}

function printResponse(response) {
  box = $("#responsebox")
  //json = jQuery.parseJSON(response)
  box.text(response)
}

$("#submitbutton").live('click', submitData)
$("#getbutton").live('click', getData)


