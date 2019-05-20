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
    {
        data: "CreatedAt",
        render: function (data, type, row, meta) {
            return data.substring(0, 10)
        },
        name: "created_at"
    },
    {
        data: "UpdatedAt",
        render: function (data, type, row, meta) {
            return data.substring(0, 10)
        },
        name: "updated_at"
    },
    {
        data: "users",
        orderable: false,
        render: function (data) {
            return data.length;

        }
    },
    {
        data: "people",
        name: "totalResponses",
        orderable: false,
        render: function (data, type, row, meta) {
            let finalResponses = 0;
            for (let i = 0; i < data.length; i++)
                if (data[i].final_response)
                    finalResponses++;
            return finalResponses + '/' + data.length;
        }
    },
    {
        data: {
            final_action: "final_action",
            users: "users"
        },
        name: "finalAction",
        orderable: false,
        render: function (data) {
            if (data.final_action.String === "") {
                return "لا";
            } else {
                return "نعم";
            }
        }
    }
];

$(document).ready(function () {

    $('#selectedUsers').select2({
        placeholder: "الاسم",
        dir: "rtl"
    });

    $('#username').select2({
        placeholder: "اسم المستخدم",
        dir: "rtl"
    });



    $(".datepicker").datepicker({
        format: 'yyyy-mm-dd'
    });
    $("#resetSearch").on('click', function () {
        $('#searchForm')[0].reset();
        $('.search').trigger("change");
        return false;
    });
    $("#czContainer").czMore();


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
            type: "textarea",
            attr: {
                disabled: !isAdmin,
                readonly: !isAdmin
            }
        }, {
            label: "الاجراء النهائي:",
            name: "final_action",
            type: "textarea",
            attr: {
                disabled: isAdmin,
                readonly: isAdmin
            }
        }],
        i18n: {
            create: {
                button: "اضافه",
                title: "اضافه تكليف",
                submit: "حفظ التكليف"
            },
            edit: {
                button: "تعديل",
                title: "تعديل تكليف",
                submit: "حفظ التكليف"
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
        initComplete: configureTableForNonAdmins,
        createdRow: function (row, data, dataIndex) {
            if (isAdmin && !data.seen) {
                $(row).children().first().addClass('unseen');
            }
            for (let i = 0; i < data.users.length; i++) {
                if (data.users[i].user_id === userID && !data.users[i].seen) {
                    $(row).children().first().addClass('unseen');
                }
            }
        },
        language: {
            url: '/source-codes/languages/datatables.language.json'
        },
        order: [[4, 'desc']],
        rowId: "ID",
        processing: true,
        serverSide: true,
        ajax: {
            url: "/tasks/getData",
            data: function (d) {
                return $.extend({}, d, {
                    "description": $("#description").val(),
                    "sent_to": $("#sent_to").val(),
                    "min_date": $("#min").val(),
                    "max_date": $("#max").val(),
                    "retrieve": $("input[name*='retrieveValues']:checked").val()
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
            {extend: "create", editor: editor, name: "createButton"},
            {extend: "edit", editor: editor, name: "editButton"},
            {extend: "remove", editor: editor, name: "removeButton"}
        ]
    });

    search();
    sendExtraFormDataAndValidate();
    showHideChild();
    showPeopleActions();
    openModalOnDoubleClick();
    redrawTableOnModalClose();
    preventModalOpeningIfNoRecordsAreFound();
});

function preventModalOpeningIfNoRecordsAreFound() {
    editor.on('preOpen', function (e, type, action) {
        const modifier = editor.modifier();
        if (action === "edit" && modifier.length < 1) {
            console.log(modifier);
            return false;
        }
    })
}

function redrawTableOnModalClose() {
    editor.on('close', function () {
        myTable.draw();
    });
}


function configureTableForNonAdmins() {
    if (!isAdmin) {
        myTable.buttons('createButton:name, removeButton:name').remove();
    }
}

function openModalOnDoubleClick() {
    $('#baseTable tbody').on('dblclick', 'tr', function () {
        myTable.rows('.selected').deselect();
        $(this).toggleClass('selected');
        myTable.rows('.selected').edit();
        $(this).toggleClass('selected');
    });
}

function showPeopleActions() {
    editor.on('open', function (e, type) {
        const modifier = editor.modifier();
        let $selectedUsers = $('#selectedUsers');
        if (isAdmin) {
            if (modifier) {
                markTaskAsSeen(modifier);
            }
        } else {
            $selectedUsers.attr('disabled', true);
            if (modifier) {
                const data = myTable.row(modifier).data();
                for (let i = 0; i < data.users.length; i++) {
                    markPersonTaskAsSeen(data, i);
                }
            }
        }

        $selectedUsers.val(null).trigger('change');
        $('#czContainer').empty();
        const selectedPeopleIDs = [];
        if (modifier) {
            const data = myTable.row(modifier).data();
            const finalAction = this.field('final_action');
            finalAction.val(data.final_action.String);
            for (let i = 1; i <= data.users.length; i++) {
                selectedPeopleIDs.push(data.users[i - 1].user.ID);
            }
            $selectedUsers.val(selectedPeopleIDs);
            $selectedUsers.trigger('change'); // Notify any JS components that the value changed

            let i;
            for (i = 1; i <= data.people.length; i++) {
                addPersonAndHisActionToModal(i, data);
            }
        }
    });
}

function addPersonAndHisActionToModal(i, data) {
    $('#btnPlus').trigger('click');
    $('#id_' + i + '_repeat').val(data.people[i - 1].ID);
    $('#name_' + i + '_repeat').val(data.people[i - 1].name);
    $('#action_' + i + '_repeat').val(data.people[i - 1].action_taken);
    $('#finalResponse_' + i + '_repeat').prop('checked', data.people[i - 1].final_response);
}

function markTaskAsSeen(modifier) {
    const data = myTable.row(modifier).data();
    if (!data.seen) {
        $.post("/tasks/seen", {
            seen: true,
            task_id: data.ID
        });
        $('tr#' + data.ID).removeClass('unseen')
    }
}

function markPersonTaskAsSeen(data, i) {
    if (data.users[i].user_id === userID && !data.users[i].seen) {
        $.post("/tasks/person/seen", {
            seen: true,
            task_id: data.ID,
            user_id: data.users[i].user_id
        });
        $('tr#' + data.ID).removeClass('unseen')
    }
}


function search() {
    $(".search").on('keyup change', function () {
        myTable.draw()
    });
}

function sendExtraFormDataAndValidate() {
    editor.on('preSubmit', function (e, data, action) {
        if (action === 'remove')
            return;

        $(".invalid-feedback").hide();

        const description = this.field('description');


        if (description.val().length === 0) {
            description.error('يجب ان يوجد تكليف')
        }

        let numberOfPeople = $('#czContainer_czMore_txtCount').val();
        for (let i = 0; i < numberOfPeople; i++) {
            if ($('#finalResponse_' + (i + 1) + '_repeat').is(':checked')) {
                let $actionTaken = $('#action_' + (i + 1) + '_repeat');
                if ($actionTaken.val() === "") {
                    $actionTaken.next().show();
                    editor.error("حدث خطأ, برجاء مراجعه البيانات");
                }
            }
        }

        if (this.inError()) {
            return false;
        }


        let selectedUsers = $('#selectedUsers').val();
        const numberOfExtraUsers = selectedUsers.length;
        data.data['totalUsers'] = numberOfExtraUsers;
        for (let i = 0; i < numberOfExtraUsers; i++) {
            data.data["users_" + i] = selectedUsers[i];
        }

        data.data['totalPeople'] = numberOfPeople;
        for (let i = 0; i < numberOfPeople; i++) {
            data.data["people_id_" + i] = $('#id_' + (i + 1) + '_repeat').val();
            data.data["people_name_" + i] = $('#name_' + (i + 1) + '_repeat').val();
            data.data["people_action_" + i] = $('#action_' + (i + 1) + '_repeat').val();
            data.data["people_finalResponse_" + i] = $('#finalResponse_' + (i + 1) + '_repeat').is(':checked');
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
