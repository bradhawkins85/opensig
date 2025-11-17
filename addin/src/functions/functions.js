(function(){
  Office.onReady(function(){
    // ready
  });

  // Ribbon button action
  function insertSignaturePlaceholder() {
    Office.context.mailbox.item.body.setAsync('[[signature:default]]', {coercionType:'html'}, function(res){
      // no-op
    });
  }

  // Event-based activation stub for compose (if enabled in manifest)
  function onNewMessageCompose(event) {
    Office.context.mailbox.item.body.setAsync('[[signature:default]]', {coercionType:'html'}, function(){
      event.completed();
    });
  }

  // Expose to global for OfficeJS
  window.insertSignaturePlaceholder = insertSignaturePlaceholder;
  window.onNewMessageCompose = onNewMessageCompose;
})();
