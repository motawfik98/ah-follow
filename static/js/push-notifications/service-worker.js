self.addEventListener('push', event => {
    const title = 'التعديلات الوزاريه';
    const options = {
        body: event.data.text(),
        icon: "/img/notification.png",
    };

    event.waitUntil(self.registration.showNotification(title, options));
});
