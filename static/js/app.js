let editor; // use a global for the submit and return data rendering in the examples
let myTable;
const maxTextLength = ($(window).width() < 991.98) ? 50 : 180;
let cols;
let basicCols = [
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
        visible: classification !== 3,
    }
];
if (classification === 3) {
    cols = basicCols;
} else {
    cols = basicCols.concat([
        {
            width: "5%",
            data: "following_users",
            orderable: false,
            render: function (data) {
                return data.length;

            },
        }, {
            width: "5%",
            data: "workingOn_users",
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
            data: "workingOn_users",
            render: function (data) {
                return generateWorkingOnTable(data);
            }
        }
    ]);
}

function alternateRecordsetColors(index) {
    let $currentRecordset = $('.recordset').eq(index - 1);
    $currentRecordset.removeClass('even odd');
    if ((index - 1) % 2 === 0)
        $currentRecordset.addClass('even');
    else
        $currentRecordset.addClass('odd');

}

$(document).ready(function () {

    $('#selectedFollowingUsers').select2({
        placeholder: "الاسم",
        dir: "rtl"
    });

    $('#username').select2({
        placeholder: "اسم المستخدم",
        dir: "rtl"
    });

    $('#sent_to').select2({
        placeholder: "اسم المستخدم",
        dir: "rtl",
        allowClear: true
    });


    $(".datepicker").datepicker({
        format: 'yyyy-mm-dd'
    });
    $("#resetSearch").on('click', function () {
        $('#searchForm')[0].reset();
        $('#sent_to').val('').trigger('change');
        $('.search').trigger("change");
        return false;
    });
    $("#czContainer").czMore({
        onAdd: function (index) {
            $('.workingOn-select2').select2({
                placeholder: 'اسم القائم به',
                dir: "rtl",
            });
            if ($('.recordset').length > 0) {
                alternateRecordsetColors(index);
            }
        }, onDelete: function (id) {
            let length = $('.recordset').length;
            if (length > 0) {
                for (let index = 1; index <= length; index++) {
                    alternateRecordsetColors(index);
                }

            }
        }
    });


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
                disabled: classification !== 1,
                readonly: classification !== 1
            }
        }, {
            label: "الاجراء النهائي:",
            name: "final_action",
            type: "textarea",
            attr: {
                disabled: classification === 1,
                readonly: classification === 1
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
            if (classification === 1 && !data.seen) {
                $(row).children().first().addClass('unseen');
            }
            if (classification === 2) {
                for (let i = 0; i < data.following_users.length; i++) {
                    if (data.following_users[i].user_id === userID) {
                        if (!data.following_users[i].seen || data.following_users[i].new_from_minister)
                            $(row).children().first().addClass('unseen');
                        else if (data.following_users[i].new_from_working_on_user)
                            $(row).children().first().addClass('new_from_working_on_user');
                        else if (data.following_users[i].marked_as_unseen)
                            $(row).children().first().addClass('marked_as_unseen');
                    }
                }
            }
            if (classification === 3) {
                for (let i = 0; i < data.workingOn_users.length; i++) {
                    if (data.workingOn_users[i].user_id === userID) {
                        if (!data.workingOn_users[i].seen)
                            $(row).children().first().addClass('unseen');
                        else if (data.workingOn_users[i].marked_as_unseen)
                            $(row).children().first().addClass('marked_as_unseen');
                    }
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
                let values = getFilteredAttributes();
                return $.extend({}, d, {
                    "description": values[0],
                    "sent_to": values[1],
                    "min_date": values[2],
                    "max_date": values[3],
                    "retrieve": values[4],
                    "hash": values[5]
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
            addDataTableButton("remove", "مسح تكليف"),
            {
                extend: 'collection',
                text: "طباعه",
                buttons: [
                    {
                        text: 'طباعه مختصره',
                        action: function (e, dt, node, config) {
                            let PDFUrl = generatePDFUrl() + "&collapsed=true";
                            window.open(PDFUrl, "_blank");
                        }
                    },
                    {
                        text: 'طباعه بالقائمين به',
                        action: function (e, dt, node, config) {
                            let PDFUrl = generatePDFUrl() + "&collapsed=false";
                            window.open(PDFUrl, "_blank");
                        }
                    }
                ],
            }
        ]
    });

    search();
    sendExtraFormDataAndValidate();
    showPeopleActions();
    openModalOnDoubleClick();
    redrawTableOnModalClose();
    preventModalOpeningIfNoRecordsAreFound();
    $('#btnRefresh').on('click', function () {
        myTable.draw();
    });
});

function generatePDFUrl() {
    let values = getFilteredAttributes();
    return `/generate-pdf?description=${values[0]}&sent_to=${values[1]}&min_date=${values[2]}&max_date=${values[3]}`
        + `&retrieve=${values[4]}&hash=${values[5]}&sort_column=${values[6]}&sort_direction=${values[7]}`
}

function getFilteredAttributes() {
    let description = $("#description").val();
    let sent_to = $("#sent_to").val();
    let min_date = $("#min").val();
    let max_date = $("#max").val();
    let retrieve = $("input[name*='retrieveValues']:checked").val();
    let hash = (getUrlVars()["hash"] !== "") ? getUrlVars()["hash"] : "";
    let order = myTable.order();
    let sort_column = order[0][0];
    let sort_direction = order[0][1];
    return [description, sent_to, min_date, max_date, retrieve, hash, sort_column, sort_direction]
}

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
    if (classification !== 1) {
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
        if (classification === 3) {
            removeFilesUpload();
        }
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
        let $selectedUsers = $('#selectedFollowingUsers');
        if (classification === 1) {
            if (modifier) {
                const data = myTable.row(modifier).data();
                if (!data.seen) {
                    changeTaskSeenProperty(data.ID, true);
                }
            }
        } else if (classification === 2) {
            $selectedUsers.attr('disabled', true);
            if (modifier) {
                const data = myTable.row(modifier).data();
                for (let i = 0; i < data.following_users.length; i++) {
                    if (data.following_users[i].user_id === userID) {
                        if (!data.following_users[i].seen || data.following_users[i].marked_as_unseen ||
                            data.following_users[i].new_from_working_on_user || data.following_users[i].new_from_minister) {
                            changeUserTaskSeenProperty(data.ID, data.following_users[i].user_id, true, true);
                        }
                    }
                }
            }
        } else {
            $selectedUsers.attr('disabled', true);
            if (modifier) {
                const data = myTable.row(modifier).data();
                for (let i = 0; i < data.workingOn_users.length; i++) {
                    if (data.workingOn_users[i].user_id === userID) {
                        if (!data.workingOn_users[i].seen || data.workingOn_users[i].marked_as_unseen) {
                            changeUserTaskSeenProperty(data.ID, data.workingOn_users[i].user_id, true, false);
                        }
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
            for (let i = 1; i <= data.following_users.length; i++) {
                selectedPeopleIDs.push(data.following_users[i - 1].user_id);
            }
            $selectedUsers.val(selectedPeopleIDs);
            $selectedUsers.trigger('change'); // Notify any JS components that the value changed

            for (let i = 1; i <= data.workingOn_users.length; i++) {
                addPersonAndHisActionToModal(i, data);
            }
            $markAsSeen = $('#markAsSeen');
            $markAsSeen.off();
            $markAsSeen.on('click', function () {
                if (classification === 1)
                    changeTaskSeenProperty(data.ID, false);
                else if (classification === 2)
                    changeUserTaskSeenProperty(data.ID, userID, false, true);
                else
                    changeUserTaskSeenProperty(data.ID, userID, false, false);
            });
        }
    });
}

function removeFilesUpload() {
    $('.eu_table .row').first().remove();
    $('.eu_table .second .drop').remove();
    $('.btn.remove').remove();

}

function addPersonAndHisActionToModal(i, data) {
    $('#btnPlus').trigger('click');
    $('#id_' + i + '_repeat').val(data.workingOn_users[i - 1].ID);
    $('#user_id_' + i + '_repeat').val(data.workingOn_users[i - 1].user_id).trigger('change').attr("readonly", true).attr("disabled", true);
    let $actionTaken = $('#action_' + i + '_repeat');
    let $workingOnNotes = $('#workingOnNotes_' + i + '_repeat');
    let $finalResponse = $('#finalResponse_' + i + '_repeat');
    $actionTaken.val(data.workingOn_users[i - 1].action_taken);
    $workingOnNotes.val(data.workingOn_users[i - 1].notes);
    $finalResponse.prop('checked', data.workingOn_users[i - 1].final_response);
    if (classification === 3) {
        $actionTaken.attr("readonly", true).attr("disabled", true);
        $finalResponse.attr("readonly", true).attr("disabled", true);
    }
}

function changeTaskSeenProperty(taskID, seenProperty) {
    $.post("/tasks/seen", {
        seen: seenProperty,
        task_id: taskID
    });

}

function changeUserTaskSeenProperty(taskID, userID, seenProperty, isFollower) {
    $.post("/tasks/person/seen", {
        seen: seenProperty,
        task_id: taskID,
        user_id: userID,
        is_follower: isFollower,
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


        let selectedUsers = $('#selectedFollowingUsers').val();
        const numberOfExtraUsers = selectedUsers.length;
        data.data['totalUsers'] = numberOfExtraUsers;
        for (let i = 0; i < numberOfExtraUsers; i++) {
            data.data["following_users_" + i] = selectedUsers[i];
        }

        data.data['totalWorkingOnUsers'] = numberOfPeople;
        for (let i = 0; i < numberOfPeople; i++) {
            data.data["people_id_" + i] = $('#id_' + (i + 1) + '_repeat').val();
            data.data["people_user_id_" + i] = $('#user_id_' + (i + 1) + '_repeat').val();
            data.data["people_action_" + i] = $('#action_' + (i + 1) + '_repeat').val();
            data.data["people_notes_" + i] = $('#workingOnNotes_' + (i + 1) + '_repeat').val();
            data.data["people_finalResponse_" + i] = $('#finalResponse_' + (i + 1) + '_repeat').is(':checked');
        }


    })
}

function generateWorkingOnTable(data) {
    let innerTable = '<table cellpadding="5" cellspacing="0" border="0" style="padding-left:50px; width: 80%">';
    innerTable += '<thead><tr>';
    innerTable += '<th>القائم به</th><th>الموقف</th>';
    innerTable += '</tr></head><tbody>';
    for (let i = 0; i < data.length; i++) {
        innerTable += '<tr>';
        innerTable += ('<td>' + data[i].user.username + '</td>');
        innerTable += ('<td>' + data[i].action_taken + '</td>');
        innerTable += '</tr>';
    }
    innerTable += '</tbody></table>';
    return innerTable;
}

function getUrlVars() {
    var vars = [], hash;
    var hashes = window.location.href.slice(window.location.href.indexOf('?') + 1).split('&');
    for (var i = 0; i < hashes.length; i++) {
        hash = hashes[i].split('=');
        vars.push(hash[0]);
        vars[hash[0]] = hash[1];
    }
    return vars;
}