
function toggleContent(ref) {
  content = $("#" + ref)
  if (content.css("display") == "none") {
    content.css("display", "block")
  } else {
    content.css("display", "none")
  }
}

