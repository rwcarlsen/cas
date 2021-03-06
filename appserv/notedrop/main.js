
function main() {
  initHandlers()
}

function initHandlers() {
  $("#submit-btn").on('click', submitNote)
}

function submitNote(eventObj) {
  metabox = $("#meta-box")
  bodybox = $("#body-box")
  meta = metabox.val()
  body = bodybox.val()

  try {
    note = jQuery.parseJSON("{" + meta + "}");
  } catch (err) {
    alert("Badly formed json in metadata.\n\nFix and submit again.")
    return
  }

  note.body = body
  data = JSON.stringify(note)
  $.post("/notedrop/putnote", data, printResponse)
}

function getData(eventObj) {
  box = $("#refbox")
  data = box.val()
  $.post("/get", data, printResponse)
}

function printResponse(response) {
  try {
    jQuery.parseJSON(response);
    msg = "Submission successful:\n\n" + response
  } catch (err) {
    msg = "Submission error: " + response
  }
  box = $("#status-box")
  box.text(msg)
}

main()
