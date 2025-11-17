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

  // Event-based activation for compose (LaunchEvent)
  function onNewMessageCompose(event) {
    // Check if mailbox item is available
    if (!Office.context.mailbox || !Office.context.mailbox.item) {
      console.error('Mailbox context not available');
      event.completed();
      return;
    }

    // Auto-insert signature placeholder at the end of the body
    // Using setAsync to append at the end instead of replacing content
    Office.context.mailbox.item.body.getAsync(
      Office.CoercionType.Html,
      function(result) {
        if (result.status === Office.AsyncResultStatus.Succeeded) {
          // Check if placeholder already exists to avoid duplicates
          const bodyContent = result.value;
          if (bodyContent.indexOf('[[signature:') === -1) {
            // Append signature placeholder at the end
            Office.context.mailbox.item.body.setAsync(
              bodyContent + '<br/><br/>[[signature:default]]',
              { coercionType: Office.CoercionType.Html },
              function(setResult) {
                if (setResult.status === Office.AsyncResultStatus.Failed) {
                  console.error('Failed to insert signature on compose:', setResult.error);
                }
                event.completed();
              }
            );
          } else {
            // Placeholder already exists, don't insert
            event.completed();
          }
        } else {
          console.error('Failed to read body:', result.error);
          event.completed();
        }
      }
    );
  }

  // Register LaunchEvent handlers
  if (Office.actions) {
    Office.actions.associate('onNewMessageCompose', onNewMessageCompose);
  }

  // Expose to global for OfficeJS (ribbon button)
  window.insertSignaturePlaceholder = insertSignaturePlaceholder;
})();
