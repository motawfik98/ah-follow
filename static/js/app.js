let editor; // use a global for the submit and return data rendering in the examples
let myTable;

let cols = [
    {
        data: "description",
        orderable: false,
        className: 'select-checkbox',
        render: function () {
            return ''
        }
    }, {
        class: "details-control",
        orderable: false,
        data: "description",
        render: function () {
            return '';
        }
    }, {data: "description", name: "description"},
    {data: "followed_by", name: "followed_by"},
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
    $("#resetSearch").on('click', function () {
        $('#searchForm')[0].reset();
        $('.search').trigger("change");
        return false;
    });
    $("#czContainer").czMore({
        onDelete: function (id) {
            $.post('/tasks/removeChild', {id: id});
            myTable.draw();
        }
    });


    editor = new $.fn.dataTable.Editor({
        table: "#baseTable",
        template: '#customForm',
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
            label: "المتابع:",
            name: "followed_by"
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


    myTable = $('#baseTable').DataTable({
        language: {
            url: '//cdn.datatables.net/plug-ins/9dcbecd42ad/i18n/Arabic.json'
        },
        order: [[4, 'desc']],
        rowId: "ID",
        processing: true,
        serverSide: true,
        ajax: {
            url: "/getData",
            data: function (d) {
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
            style: 'os',
            selector: 'td:first-child'
        },     // enable single row selection
        buttons: [
            {extend: "create", editor: editor},
            {extend: "edit", editor: editor},
            {extend: "remove", editor: editor}
        ]
    });

    search();
    sendExtraFormData();
    showHideChild();
    showPeopleActions();
});


function showPeopleActions() {
    editor.on('open', function (e, type) {
        $('#czContainer').empty();
        const modifier = editor.modifier();

        if (modifier) {
            const data = myTable.row(modifier).data();
            for (let i = 1; i <= data.people.length; i++) {
                $('#btnPlus').trigger('click');
                $('#id_' + i + '_repeat').val(data.people[i - 1].ID);
                $('#name_' + i + '_repeat').val(data.people[i - 1].name);
                $('#action_' + i + '_repeat').val(data.people[i - 1].action_taken);
            }
        }
    });
}


function search() {
    $(".search").on('keyup change', function () {
        myTable.draw()
    });
}

function sendExtraFormData() {
    editor.on('preSubmit', function (e, data, action) {
        if (action === 'remove')
            return;
        const numberOfExtraFields = $('#czContainer_czMore_txtCount').val();
        for (let i = 1; i <= numberOfExtraFields; i += 1) {
            let firstInputName = 'name_' + i + '_repeat';
            let secondInputName = 'action_' + i + '_repeat';
            data.data['totalPeople'] = numberOfExtraFields;
            data.data[firstInputName] = $('#' + firstInputName).val();
            data.data[secondInputName] = $('#' + secondInputName).val();
        }
    })
}


function showHideChild() {
    // Array to track the ids of the details displayed rows
    const detailRows = [];

    $('#baseTable tbody').on('click', 'tr td.details-control', function () {
        const tr = $(this).closest('tr');
        const row = myTable.row(tr);
        const idx = $.inArray(tr.attr('id'), detailRows);

        if (row.child.isShown()) {
            tr.removeClass('details');
            row.child.hide();

            // Remove from the 'open' array
            detailRows.splice(idx, 1);
        } else {
            tr.addClass('details');
            row.child(format(row.data())).show();

            // Add to the 'open' array
            if (idx === -1) {
                detailRows.push(tr.attr('id'));
            }
        }
    });

    // On each draw, loop over the `detailRows` array and show any child rows
    myTable.on('draw', function () {
        $.each(detailRows, function (i, id) {
            $('#' + id + ' td.details-control').trigger('click');
        });
    });
}

function format(d) {
    let innerTable = '<table cellpadding="5" cellspacing="0" border="0" style="padding-left:50px; width: 80%">';
    innerTable += '<thead><tr>';
    innerTable += '<th>القائم به</th><th>الموقف</th>';
    innerTable += '</tr></head><tbody>';
    for (var i = 0; i < d.people.length; i++) {
        innerTable += '<tr>';
        innerTable += ('<td>' + d.people[i].name + '</td>');
        innerTable += ('<td>' + d.people[i].action_taken + '</td>');
        innerTable += '</tr>';
    }
    innerTable += '</tbody></table>';
    return innerTable;
}
