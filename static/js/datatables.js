var editor; // use a global for the submit and return data rendering in the examples

$(document).ready(function () {
    editor = new $.fn.dataTable.Editor({
        table: "#baseTable",
        fields: [{
            label: "التكليف:",
            name: "description",
            type: "textarea"
        }, {
            label: "القائم به:",
            name: "sentTo",
            type: "textarea",
        }, {
            label: "المتابع:",
            name: "followedBy"
        }, {
            label: "الموقف",
            name: "actionTaken"
        }],
        i18n: {
            create: {
                button: "اضافه",
                title: "اضافه تكليف",
                submit: "اضافه"
            },
            edit: {
                button: "تعديل",
                title: "تعديل تكليف",
                submit: "تعديل"
            },
            remove: {
                button: "مسح",
                title: "مسح تكليف",
                submit: "مسح",
                confirm: {
                    _: "هل انت متأكد من مسح %d سجل ؟",
                    1: "هل انت متأكد من مسح هذا سجل ؟"
                }
            }
        }
    });


    $('#baseTable').DataTable({
        language: {
            url: '//cdn.datatables.net/plug-ins/9dcbecd42ad/i18n/Arabic.json'
        },
        dom: 'Bfrtip',        // element order: NEEDS BUTTON CONTAINER (B) ****
        select: true,     // enable single row selection
        buttons: [
            {extend: "create", editor: editor},
            {extend: "edit", editor: editor},
            {extend: "remove", editor: editor }
        ]
    });
});