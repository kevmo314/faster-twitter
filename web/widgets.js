(function () {
  // replicate the behavior of the twitter widgets.js.
  // first, find the blockquote element that is the previous sibling of the script element.
  var script = document.currentScript;
  var blockquote = script.previousElementSibling;

  // ensure it has the correct class name
  if (blockquote.className !== "twitter-tweet") {
    return;
  }

  // replace it with an iframe pointing at /embed/<tweet-id>
  var href = blockquote.lastElementChild.href;
  console.log(href);
  var tweetId = href.split("?")[0].split("/").pop();
  var iframe = document.createElement("iframe");
  iframe.src = "/embed/" + tweetId;
  iframe.width = "100%";
  iframe.allowTransparency = "true";
  iframe.allowFullscreen = "true";
  iframe.scrolling = "no";
  iframe.style = "border: none; overflow: hidden; max-width: 550px;";
  iframe.onload = function () {
    iframe.width = iframe.contentWindow.document.body.scrollWidth;
    iframe.height = iframe.contentWindow.document.body.scrollHeight;
  };
  blockquote.parentNode.replaceChild(iframe, blockquote);
})();
