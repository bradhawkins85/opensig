(function(){
  Office.onReady(function(){
    // ready
  });

  // Ribbon button action - inserts default signature
  function insertSignaturePlaceholder(event) {
    Office.context.mailbox.item.body.setSelectedDataAsync(
      '[[signature:default]]',
      { coercionType: Office.CoercionType.Html },
      function(result) {
        if (result.status === Office.AsyncResultStatus.Failed) {
          console.error('Failed to insert signature:', result.error);
        }
        // Complete the event if provided (for add-in commands)
        if (event) {
          event.completed();
        }
      }
    );
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
