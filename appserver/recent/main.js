
function toggleContent(ref) {
  content = $("#" + ref)
  if (content.css("visibility") == "hidden") {
    content.css("visibility", "visible")
  } else {
    content.css("visibility", "hidden")
  }
}

