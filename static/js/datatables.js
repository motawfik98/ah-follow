let editor; // use a global for the submit and return data rendering in the examples

let cols = [
    {
        data: "description",
        orderable: false,
        className: 'select-checkbox',
        render: function () {
            return ''
        }
    },
    {data: "description", name: "description"},
    {data: "sent_to", name: "sent_to"},
    {data: "followed_by", name: "followed_by"},
    {data: "action_taken", name: "action_taken"},
    {
        data: "CreatedAt",
        render: function (data, type, row, meta) {
            return data.substring(0, 10)
        },
        name: "created_at"
    }
];

$(document).ready(function () {
    $(".datepicker").datepicker({
        format: 'yyyy-mm-dd'
    });
    $("#resetSearch").on('click', function() {
        $('#searchForm')[0].reset();
        $('.search').trigger("change");
        return false;
    });
    editor = new $.fn.dataTable.Editor({
        table: "#baseTable",
        ajax: {
            create: {
                type: 'POST',
                url: '/tasks/add'
            }, edit: {
                type: 'POST',
                url: '/tasks/edit'
            }, remove: {
                type: 'POST',
                url: '/tasks/remove'
            }
        },
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


    let myTable = $('#baseTable').DataTable({
        language: {
            url: '//cdn.datatables.net/plug-ins/9dcbecd42ad/i18n/Arabic.json'
        },
        order: [[5, 'asc']],
        rowId: "ID",
        processing: true,
        serverSide: true,
        ajax: {
            url: "/getData",
            data: function(d) {
                return $.extend({}, d, {
                    "description": $("#description").val(),
                    "followed_by": $("#followed_by").val(),
                    "min_date": $("#min").val(),
                    "max_date": $("#max").val()
                });
            }
        },
        columns: cols,
        dom: 'Brtip',        // element order: NEEDS BUTTON CONTAINER (B) ****
        select: {
            style:    'os',
            selector: 'td:first-child'
        },     // enable single row selection
        buttons: [
            {extend: "create", editor: editor},
            {extend: "edit", editor: editor},
            {extend: "remove", editor: editor}
        ]
    });
    search(myTable);
});

function search(table) {
    $(".search").on('keyup change', function () {
        table.draw()
    });

}
