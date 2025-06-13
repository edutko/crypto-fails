const buttons = document.getElementsByTagName('button');
for (let i = 0; i < buttons.length; i++) {
    const b = buttons.item(i);
    if (b.getAttribute("id")?.startsWith("copy-")) {
        const id = b.getAttribute("id").substring(5);
        b.addEventListener('click', async () => {
            const d = document.getElementById(`data-${id}`);
            if (d) {
                await navigator.clipboard.write([
                    new ClipboardItem({['text/plain']: d.value || ''})
                ]);
            }
        });
    }
}
