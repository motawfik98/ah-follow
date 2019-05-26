navigator.serviceWorker.register('service-worker.js?isAdmin=' + encodeURIComponent(isAdmin));

function urlBase64ToUint8Array(base64String) {
    const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
    const base64 = (base64String + padding)
        .replace(/\-/g, '+')
        .replace(/_/g, '/');
    const rawData = window.atob(base64);
    return Uint8Array.from([...rawData].map(char => char.charCodeAt(0)));
}

async function requestNotificationPermission() {
    // value of permission can be 'granted', 'default', 'denied'
    // granted: user has accepted the request
    // default: user has dismissed the notification permission popup by clicking on x
    // denied: user has denied the request.
    return await window.Notification.requestPermission();
}

navigator.serviceWorker.ready
    .then(function (registration) {
        return registration.pushManager.getSubscription()
            .then(async function (subscription) {
                console.log(JSON.stringify(subscription));
                if (subscription) {
                    console.log("subscribed");
                    return null;
                }
                console.log("No subscription available");
                const permission = await requestNotificationPermission();
                if (permission === "granted") {
                    const vapidPublicKey = 'BHdQL2HMczQYoKR7EIlGBaUSHUWrDQokRducAdSFAej7nbix6H7F00PiKT3Z0wJ4NLRSxgeRfgsPUD8-X77iLO4';
                    const convertedVapidKey = urlBase64ToUint8Array(vapidPublicKey);

                    return registration.pushManager.subscribe({
                        userVisibleOnly: true,
                        applicationServerKey: convertedVapidKey
                    });
                }
            });
    }).then(function (subscription) {
    if (subscription) {
        let jsonSubscription = JSON.parse(JSON.stringify(subscription));

        $.post('/notifications/register', {
            endpoint: jsonSubscription.endpoint,
            auth: jsonSubscription.keys.auth,
            p256dh: jsonSubscription.keys.p256dh,
        });
    }
});