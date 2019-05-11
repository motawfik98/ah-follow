let editor; // use a global for the submit and return data rendering in the examples

let cols = [
    {data: "description", name: "description"},
    {data: "sent_to", name:"sent_to"},
    {data: "followed_by", name: "followed_by"},
    {data: "action_taken", name:"action_taken"},
    {
        data: "CreatedAt",
        render: function(data, type, row, meta) {
            return data.substring(0, 10)
        },
        name: "created_at"
    }
];

$(document).ready(function () {
    editor = new $.fn.dataTable.Editor({
        table: "#baseTable",
        ajax: "/tasksHandler",
        idSrc: "ID",
        legacyAjax: true,
        fields: [{
            label: "التكليف:",
            name: "description",
            type: "textarea"
        }, {
            label: "القائم به:",
            name: "sent_to",
            type: "textarea",
        }, {
            label: "المتابع:",
            name: "followed_by"
        }, {
            label: "الموقف",
            name: "action_taken"
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
        rowId: "ID",
        processing: true,
        serverSide: true,
        ajax: "/getData",
        columns: cols,
        dom: 'Bfrtip',        // element order: NEEDS BUTTON CONTAINER (B) ****
        select: {style: 'single'},     // enable single row selection
        buttons: [
            {extend: "create", editor: editor},
            {extend: "edit", editor: editor},
            {extend: "remove", editor: editor }
        ]
    });
});