let editor; // use a global for the submit and return data rendering in the examples
let myTable;
const maxTextLength = ($(window).width() < 991.98) ? 50 : 180;
let cols = [
    {
        width: "5%",
        class: "control",
        orderable: false,
        data: "description",
        render: function () {
            return '';
        }
    }, {
        width: "5%",
        data: "description",
        orderable: false,
        className: 'select-checkbox',
        render: function () {
            return ''
        },
    }, {
        data: "description",
        name: "description",
        render: function (data, type) {
            return type === 'display' && data.length > maxTextLength ?
                "<div class='text-wrap'>" + data.substr(0, maxTextLength) + "...</div>" :
                "<div class='text-wrap'>" + data + "</div>";
        },
    }, {
        width: "10%",
        data: "CreatedAt",
        render: function (data, type, row, meta) {
            return data.substring(0, 10)
        },
        name: "created_at",
    }, {
        width: "10%",
        data: "UpdatedAt",
        render: function (data, type, row, meta) {
            return data.substring(0, 10)
        },
        name: "updated_at",
    }, {
        width: "5%",
        data: "users",
        orderable: false,
        render: function (data) {
            return data.length;

        },
    }, {
        width: "5%",
        data: "people",
        name: "totalResponses",
        orderable: false,
        render: function (data, type, row, meta) {
            let finalResponses = 0;
            for (let i = 0; i < data.length; i++)
                if (data[i].final_response)
                    finalResponses++;
            return finalResponses + '/' + data.length;
        },
    }, {
        width: "5%",
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
        },
    }, {
        width: "5%",
        orderable: false,
        name: "fullDescription",
        data: "description",
        render: function (data) {
            return data;
        }
    }, {
        width: "5%",
        orderable: false,
        name: "peopleActions",
        data: "people",
        render: function (data) {
            return generatePeopleTable(data);
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
        formOptions: {
            main: {
                onEsc: false
            }
        },
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
        }, {
            label: "ملفات:",
            ajax: '/tasks/validate-image',
            name: "files[].id",
            type: "uploadMany",
            display: function (file_id) {
                let file = editor.file('files', file_id);
                return '<a class="file-display mx-5" target="_blank" href=' + file.web_path + '>'
                    + file.created_at + '</a>';
            },
            noFileText: 'لا يوجد ملف',
            fileReadText: 'يتم الرفع',
            uploadText: 'رفع ملف',
            clearText: 'مسح الاختيار',
            dragDropText: 'اسحب الملف الى هنا ليتم الرفع',
            processingText: 'يتم الرفع'
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
        responsive: {
            details: {
                type: 'column',
                target: 'td:first-child',
                renderer: function (api, rowIdx, columns) {
                    let data = $.map(columns, function (col, i) {
                        let finalTable = "";
                        if (col.title === 'hidden people') {
                            finalTable += '<tr data-dt-row="' + col.rowIndex + '" data-dt-column="' + col.columnIndex + '">' +
                                '<td colspan="2">' + col.data + '</td>' +
                                '</tr>'
                        } else if (col.title === 'full description') {
                            if (col.data.length > maxTextLength) {
                                finalTable += '<tr data-dt-row="' + col.rowIndex + '" data-dt-column="' + col.columnIndex + '">' +
                                    '<td>التكليف:</td> ' +
                                    '<td>' + col.data + '</td>' +
                                    '</tr>'
                            }
                        } else {
                            finalTable += col.hidden ?
                                '<tr data-dt-row="' + col.rowIndex + '" data-dt-column="' + col.columnIndex + '">' +
                                '<td>' + col.title + ':' + '</td> ' +
                                '<td>' + col.data + '</td>' +
                                '</tr>'
                                :
                                '';
                        }

                        return finalTable;
                    }).join('');

                    return data ?
                        $('<table/>').append(data) :
                        false;
                }
            }
        },
        initComplete: configureTableForNonAdmins,
        createdRow: function (row, data, dataIndex) {
            let filesIDs = [];
            for (let i = 0; i < data.files.length; i++) {
                filesIDs.push({"id": data.files[i].ID.toString()})
            }
            data.files = filesIDs;
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
            selector: 'td:nth-child(2)'
        },     // enable single row selection
        buttons: [
            addDataTableButton("create", "اضافه تكليف"),
            addDataTableButton("edit", "تعديل تكليف"),
            addDataTableButton("remove", "مسح تكليف")
        ]
    });

    search();
    sendExtraFormDataAndValidate();
    showPeopleActions();
    openModalOnDoubleClick();
    redrawTableOnModalClose();
    preventModalOpeningIfNoRecordsAreFound();
});

function addDataTableButton(baseButton, text) {
    return {
        extend: baseButton,
        editor: editor,
        name: baseButton + "Button",
        formButtons: [
            {
                text: text,
                className: 'btn-primary mr-3',
                action: function () {
                    this.submit();
                }
            },
            {
                text: 'الرجوع',
                className: 'btn-secondary',
                action: function () {
                    this.close();
                }
            }
        ]
    }
}

function preventModalOpeningIfNoRecordsAreFound() {
    editor.on('preOpen', function (e, type, action) {
        const modifier = editor.modifier();
        if (action === "edit" && modifier.length < 1) {
            return false;
        }
    });
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
    editor.on('open', function (e, type, action) {
        let $markAsSeen = $('#markAsSeen');
        if (action === 'edit') {
            if ($markAsSeen.length < 1) {
                let markTaskAsUnseen = '<button class="btn btn-link mr-5" id="markAsSeen">اعتباره جديد</button>';
                $(markTaskAsUnseen).insertBefore('.DTE_Header>.close');
            } else {
                $markAsSeen.removeClass("d-none");
            }
        } else {
            $markAsSeen.addClass("d-none");
        }
        const modifier = editor.modifier();
        let $selectedUsers = $('#selectedUsers');
        if (isAdmin) {
            if (modifier) {
                const data = myTable.row(modifier).data();
                if (!data.seen) {
                    changeTaskSeenProperty(data.ID, true);
                }
            }
        } else {
            $selectedUsers.attr('disabled', true);
            if (modifier) {
                const data = myTable.row(modifier).data();
                for (let i = 0; i < data.users.length; i++) {
                    if (data.users[i].user_id === userID && !data.users[i].seen) {
                        changePersonTaskSeenProperty(data.ID, data.users[i].user_id, true);
                    }
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

            for (let i = 1; i <= data.people.length; i++) {
                addPersonAndHisActionToModal(i, data);
            }

            $markAsSeen.on('click', function () {
                if (isAdmin)
                    changeTaskSeenProperty(data.ID, false);
                else
                    changePersonTaskSeenProperty(data.ID, userID, false);

            });
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

function changeTaskSeenProperty(taskID, seenProperty) {
    $.post("/tasks/seen", {
        seen: seenProperty,
        task_id: taskID
    });

}

function changePersonTaskSeenProperty(taskID, userID, seenProperty) {
    $.post("/tasks/person/seen", {
        seen: seenProperty,
        task_id: taskID,
        user_id: userID
    });
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
            description.error('يجب ان يوجد تكليف');
            editor.error("حدث خطأ, برجاء مراجعه البيانات");
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

function generatePeopleTable(data) {
    let innerTable = '<table cellpadding="5" cellspacing="0" border="0" style="padding-left:50px; width: 80%">';
    innerTable += '<thead><tr>';
    innerTable += '<th>القائم به</th><th>الموقف</th>';
    innerTable += '</tr></head><tbody>';
    for (let i = 0; i < data.length; i++) {
        innerTable += '<tr>';
        innerTable += ('<td>' + data[i].name + '</td>');
        innerTable += ('<td>' + data[i].action_taken + '</td>');
        innerTable += '</tr>';
    }
    innerTable += '</tbody></table>';
    return innerTable;
}

