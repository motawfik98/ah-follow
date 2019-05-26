self.addEventListener('push', event => {
    const isAdmin = new URL(location.toString()).searchParams.get('isAdmin');
    event.waitUntil(
        fetch(isAdmin)
    );
    const title = 'التعديلات الوزاريه';
    const data = event.data.text().split("\n");
    const options = {
        body: data[0],
        icon: "/img/notification.png",
    };

    if ((isAdmin && data[1] === "false") || (!isAdmin && data[1] === "true")) {
        event.waitUntil(self.registration.showNotification(title, options));
    }
});
