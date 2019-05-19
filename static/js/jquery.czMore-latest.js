/*
Title: Cozeit More plugin by Yasir Atabani
Documentation: na
Author: Yasir O. Atabani
Website: http://www.cozeit.com
Twitter: @yatabani

MIT License, https://github.com/cozeit/czMore/blob/master/LICENSE.md
*/
(function ($, undefined) {
    $.fn.czMore = function (options) {

        //Set defauls for the control
        var defaults = {
            max: 10,
            min: 0,
            onLoad: null,
            onAdd: null,
            onDelete: null,
            styleOverride: false,
        };
        //Update unset options with defaults if needed
        var options = $.extend(defaults, options);
        $(this).bind("onAdd", function (event, data) {
            options.onAdd.call(event, data);
        });
        $(this).bind("onLoad", function (event, data) {
            options.onLoad.call(event, data);
        });
        $(this).bind("onDelete", function (event, data) {
            options.onDelete.call(event, data);
        });
        //Executing functionality on all selected elements
        return this.each(function () {
            var obj = $(this);
            var i = obj.children(".recordset").length;
            var divPlus = '<div id="btnPlus" class="btnPlus"/>';
            var count = '<input id="' + this.id + '_czMore_txtCount" name="' + this.id + '_czMore_txtCount" type="hidden" value="0" size="5" />';

            obj.before(count);
            var recordset = obj.children("#first");
            obj.after(divPlus);
            var set = recordset.children(".recordset").children().first();
            var btnPlus = obj.siblings("#btnPlus");

            if (!options.styleOverride) {
                btnPlus.css({
                    'float': 'right',
                    'border': '0px',
                    'background-image': 'url("img/add.png")',
                    'background-position': 'center center',
                    'background-repeat': 'no-repeat',
                    'height': '25px',
                    'width': '25px',
                    'cursor': 'pointer',
                });
            }

            if (recordset.length) {
                obj.siblings("#btnPlus").click(function () { // custom code to check for empty names before adding another row
                    let error = false;
                    $('#czContainer').children().each(function () {
                        if ($(this).length !== 0) {
                            const $formRow = $(this).children().last();
                            const $firstCol = $formRow.children().first();
                            const $input = $firstCol.children().last();
                            if ($.trim($input.val()) === '') {
                                error = true;
                            }
                        }
                    });
                    if (error) {
                        return;
                    }


                    var i = obj.children(".recordset").length;
                    var item = recordset.clone().html();
                    i++;
                    item = item.replace(/\[([0-9]\d{0})\]/g, "[" + i + "]");
                    item = item.replace(/\_([0-9]\d{0})\_/g, "_" + i + "_");
                    //$(element).html(item);
                    //item = $(item).children().first();
                    //item = $(item).parent();

                    obj.append(item);
                    loadMinus(obj.children().last());
                    minusClick(obj.children().last());
                    if (options.onAdd != null) {
                        obj.trigger("onAdd", i);
                    }

                    obj.siblings("input[name$='czMore_txtCount']").val(i);

                    $(".modal").animate({scrollTop: $(".modal").height()}, "slow");


                    return false;
                });
                recordset.remove();
                for (var j = 0; j <= i; j++) {
                    loadMinus(obj.children()[j]);
                    minusClick(obj.children()[j]);
                    if (options.onAdd != null) {
                        obj.trigger("onAdd", j);
                    }
                }

                if (options.onLoad != null) {
                    obj.trigger("onLoad", i);
                }
                //obj.bind("onAdd", function (event, data) {
                //If you had passed anything in your trigger function, you can grab it using the second parameter in the callback function.
                //});
            }

            function resetNumbering() {
                $(obj).children(".recordset").each(function (index, element) {
                    $(element).find('input:text, input:password, input:hidden, input:file, select, textarea').each(function () {
                        old_name = this.name;
                        new_name = old_name.replace(/\_([0-9]\d{0})\_/g, "_" + (index + 1) + "_");
                        this.id = this.name = new_name;
                        //alert(this.name);
                    });
                    index++
                    minusClick(element);
                });
            }

            function loadMinus(recordset) {
                var divMinus = '<div id="btnMinus" class="btnMinus" />';
                $(recordset).children().first().before(divMinus);
                var btnMinus = $(recordset).children("#btnMinus");
                if (isAdmin) {
                    if (!options.styleOverride) {
                        btnMinus.css({
                            'float': 'right',
                            'border': '0px',
                            'background-image': 'url("img/remove.png")',
                            'background-position': 'center center',
                            'background-repeat': 'no-repeat',
                            'height': '25px',
                            'width': '25px',
                            'cursor': 'poitnter',
                        });
                    }
                }
            }

            function minusClick(recordset) {
                $(recordset).children("#btnMinus").click(function () {
                    var i = obj.children(".recordset").length;
                    var id = $(recordset).children().last().children().first().children().first().val();
                    $(recordset).remove();
                    resetNumbering();
                    obj.siblings("input[name$='czMore_txtCount']").val(obj.children(".recordset").length);
                    i--;
                    if (options.onDelete != null) {
                        // if (id != null)
                        obj.trigger("onDelete", id);
                    }
                });
            }
        });
    };
})(jQuery);
