function compareJSON() {
    const left = JSON.parse(leftEditor.getValue());
    const right = JSON.parse(rightEditor.getValue());
  
    const delta = jsondiffpatch.diff(left, right);
    const html = jsondiffpatch.formatters.html.format(delta, left);
  
    document.getElementById("diffOutput").innerHTML = html;
  }
  