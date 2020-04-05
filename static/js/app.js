let myTable;
const maxTextLength = ($(window).width() < 991.98) ? 50 : 180;
let cols;
let deletedFiles = [];
let allUploadedFiles = [];
let newFilesNames = new Map();
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
        render: function (data, type, row, meta) {
            let timestamp = Math.random().toString(36).substring(7);
            let editLink = "/tasks/task/" + row.Hash + "/" + timestamp;
            return '<a href="' + editLink + '">' + 'تعديل' + '</a>';
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

    let czContainer = $("#czContainer");
    if (czContainer.length > 0) {
        czContainer.czMore({
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
    }

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
            $(row).children().first().addClass(data.seen_status);
        },
        language: {
            url: '/static/source-codes/languages/datatables.language.json'
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
        buttons: [
            {
                text: "تكليف جديد",
                name: "createButton",
                action: function () {
                    let timestamp = Math.random().toString(36).substring(7);
                    window.location.replace("/tasks/task/new/" + timestamp);
                }
            },
            {
                extend: 'collection',
                text: "طباعه التكليفات",
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
    $('#btnRefresh').on('click', function () {
        myTable.draw();
    });
    configureFileUpload();
    changeFileName();
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


function configureTableForNonAdmins() {
    if (classification !== 1) {
        myTable.buttons('createButton:name').remove();
    }
}


function
showPeopleActions() {
    let $markAsSeen = $('#markAsSeen');
    $markAsSeen.on('click', function () {
        $.post("mark-as-unseen", {
            task_id: task.ID,
        }, function (data) {
            $("#seen-feedback").addClass(data["status"]).text(data["message"]);
        });
    });

    let $selectedUsers = $('#selectedFollowingUsers');
    if (classification !== 1) {
        $selectedUsers.attr('disabled', true);
    }

    $selectedUsers.val(null).trigger('change');
    $('#czContainer').empty();
    const selectedPeopleIDs = [];

    if (task.following_users !== null) {
        for (let i = 1; i <= task.following_users.length; i++) {
            selectedPeopleIDs.push(task.following_users[i - 1].user_id);
        }
        $selectedUsers.val(selectedPeopleIDs);
        $selectedUsers.trigger('change'); // Notify any JS components that the value changed
    }

    if (task.workingOn_users !== null) {
        for (let i = 1; i <= task.workingOn_users.length; i++) {
            addPersonAndHisActionToForm(i, task);
        }
    }
}

function addPersonAndHisActionToForm(i, data) {
    $('#btnPlus').trigger('click');
    $('#id_' + i + '_repeat').val(data.workingOn_users[i - 1].ID);
    $('#user_id_' + i + '_repeat').val(data.workingOn_users[i - 1].user_id).trigger('change').attr("readonly", true).attr("disabled", true);
    let $actionTaken = $('#action_' + i + '_repeat');
    let $workingOnNotes = $('#workingOnNotes_' + i + '_repeat');
    let $finalResponse = $('#finalResponse_' + i + '_repeat');
    $actionTaken.val(data.workingOn_users[i - 1].action_taken);
    $workingOnNotes.val(data.workingOn_users[i - 1].notes);
    $finalResponse.prop('checked', data.workingOn_users[i - 1].final_response);
    if (classification !== 2) {
        $actionTaken.attr("readonly", true).attr("disabled", true);
        $finalResponse.attr("readonly", true).attr("disabled", true);
    }
}

function search() {
    $(".search").on('keyup change', function () {
        myTable.draw()
    });
}

function sendExtraFormDataAndValidate() {
    $('#addEditForm').on('submit', function (e) {

        $(".invalid-feedback").hide();
        const description = $('#description');
        let errorFound = false;
        if (description.val().length === 0) {
            errorFound = true;
            description.next().show();
        }

        let numberOfPeople = $('#czContainer_czMore_txtCount').val();
        for (let i = 0; i < numberOfPeople; i++) {
            let $actionTaken = $('#action_' + (i + 1) + '_repeat');
            let $finalResponse = $('#finalResponse_' + (i + 1) + '_repeat');
            let $userID = $('#user_id_' + (i + 1) + '_repeat');
            if ($actionTaken.val() === "" && $finalResponse.is(':checked')) {
                $actionTaken.next().show();
                errorFound = true;
            }
            if ($actionTaken.val() !== "" && $userID.val() === "") {
                $userID.next().next().show();
                errorFound = true;

            }
        }

        if (errorFound) {
            e.preventDefault();
            e.stopPropagation();
            $('#form-validation').show();
            return;
        }

        let formData = new FormData($('#addEditForm')[0]);

        formData.append("id", task.ID + "");

        let selectedUsers = $('#selectedFollowingUsers').val();
        const numberOfExtraUsers = selectedUsers.length;
        formData.append("totalUsers", numberOfExtraUsers);
        for (let i = 0; i < numberOfExtraUsers; i++) {
            formData.append("following_users_" + i, selectedUsers[i])
        }

        formData.append("totalWorkingOnUsers", numberOfPeople);
        for (let i = 0; i < numberOfPeople; i++) {
            formData.append("people_id_" + i, $('#id_' + (i + 1) + '_repeat').val());
            formData.append("people_user_id_" + i, $('#user_id_' + (i + 1) + '_repeat').val());
            formData.append("people_action_" + i, $('#action_' + (i + 1) + '_repeat').val());
            formData.append("people_notes_" + i, $('#workingOnNotes_' + (i + 1) + '_repeat').val());
            formData.append("people_finalResponse_" + i, $('#finalResponse_' + (i + 1) + '_repeat').prop('checked'));
        }

        let numberOfDeletedFiles = deletedFiles.length;
        formData.append("totalDeletedFiles", numberOfDeletedFiles + "");
        for (let i = 0; i < numberOfDeletedFiles; i++) {
            formData.append("deleted_file_" + i, deletedFiles[i]);
        }
        formData.delete("files");

        for (let i = 0; i < allUploadedFiles.length; i++) {
            formData.append('files', allUploadedFiles[i]);
        }

        formData.append("totalRenamedFiles", newFilesNames.size + "");
        let fileNumber = 0;
        for (let [fileHash, newFileName] of newFilesNames) {
            formData.append("file_hash_" + fileNumber, fileHash);
            formData.append("file_name_" + fileNumber, newFileName);
            fileNumber++;
        }

        e.preventDefault();
        let actionUrl = $(this).attr('action');
        $.ajax({
            method: 'post',
            processData: false,
            contentType: false,
            cache: false,
            data: formData,
            url: actionUrl,
            beforeSend: function () {
                let $submitButton = $('#btn-form-submit');
                $submitButton.attr('disabled', 'true');
                $submitButton.html(`<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span>جاري الحفظ`);
            },
            success: function (data) {
                window.location = "/";
            },
            failure: function () {
                let $submitButton = $('#btn-form-submit');
                $submitButton.attr('disabled', 'false');
                $submitButton.html(`تعديل`);
            }
        });
    });
}

function generateWorkingOnTable(data) {
    let innerTable = '<table cellpadding="5" cellspacing="0" border="0" style="padding-left:50px; width: 80%">';
    innerTable += '<thead><tr>';
    innerTable += '<th width="20%">القائم به</th><th>الموقف</th>';
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

function configureFileUpload() {
    deletedFiles = [];
    allUploadedFiles = [];
    let $customFileInput = $('#customFile');
    if ($customFileInput.length === 0)
        return;
    $customFileInput.on('change', function () {
        let $newFilesList = $('#new-files');
        let files = $customFileInput.prop('files');
        allUploadedFiles.push.apply(allUploadedFiles, files);
        viewFilesToBeUploaded($newFilesList);
    });
    $('#display-old-files').on('click', '.file-deletion', function () {
        if ($(this).hasClass('delete')) {
            deleteFile($(this));
        } else {
            restoreFile($(this));
        }
    });

    $('#display-new-files').on('click', '.cancel', function () {
        let fileIndex = $(this).parent().index();
        allUploadedFiles.splice(fileIndex, 1);
        viewFilesToBeUploaded($('#new-files'));
    });
}

function deleteFile($fileClicked) {
    let hash = $fileClicked.parent().next().val();
    deletedFiles.push(hash);
    $fileClicked.parent().parent().addClass('list-group-item-danger');
    $fileClicked.removeClass('delete');
    $fileClicked.addClass('restore');
    $fileClicked.text('استرجاع');
}

function restoreFile($fileClicked) {
    let hash = $fileClicked.parent().next().val();
    let fileIndex = deletedFiles.indexOf(hash);
    if (fileIndex > -1)
        deletedFiles.splice(fileIndex, 1);
    $fileClicked.parent().parent().removeClass('list-group-item-danger');
    $fileClicked.addClass('delete');
    $fileClicked.removeClass('restore');
    $fileClicked.text('مسح');
}

function viewFilesToBeUploaded($newFilesList) {
    $newFilesList.children().remove();
    for (let i = 0; i < allUploadedFiles.length; i++) {
        let fileText = "الاسم " + allUploadedFiles[i].name;
        $newFilesList.append('<li class="list-group-item d-flex justify-content-between align-items-center">' + fileText
            + '<span class="badge badge-danger badge-pill cancel" style="cursor:pointer">مسح</span></li>');
    }
}

function changeFileName() {
    let $fileNameParagraph = "";
    let originalFileName = "";
    let fileHash = "";
    $('.file-edit').on('click', function () {
        fileHash = $(this).parent().next().val();
        let fileName = $(this).prev().val();
        originalFileName = fileName;
        if (newFilesNames.has(fileHash))
            fileName = newFilesNames.get(fileHash);
        $fileNameParagraph = $(this).parent().next().next();
        $('#file-edit-name').val(fileName);
        $('#file-edit-hash').val(fileHash);
    });

    $('#btn-edit-file').on('click', function () {
        let newFileName = $('#file-edit-name').val();
        let fileHash = $('#file-edit-hash').val();
        newFilesNames.set(fileHash, newFileName);

        let paragraphText = "سوف يتم تغيير اسم الملف عند التعديل الى " + newFileName;
        $fileNameParagraph.text(paragraphText);
    });

    $('#btn-restore-name').on('click', function () {
        if (newFilesNames.has(fileHash))
            newFilesNames.delete(fileHash);
        $fileNameParagraph.text("");
    });
}
