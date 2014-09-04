var FilesView = {
    init: function(sid, aid, pid, uid, tipe) {
        $(function() {
            $.getJSON('projects', function(data) {
                if (not(data['projects'])) {
                    return;
                }
                FilesView.buidDropdown('project', pid, 'assignmentsview', data['projects']);
                $.getJSON('users', function(data) {
                    if (not(data['users'])) {
                        return;
                    }
                    FilesView.buidDropdown('user', uid, 'assignmentsview', data['users']);
                    var id = tipe === 'user' ? uid : pid;
                    $.getJSON('assignments?' + tipe + '-id=' + id, function(data) {
                        if (not(data['assignments'])) {
                            return;
                        }
                        FilesView.buidDropdown('assignment', aid, 'submissionsview', data['assignments']);
                        $.getJSON('submissions?assignment-id=' + aid, function(data) {
                            if (not(data['submissions'])) {
                                return;
                            }
                            FilesView.subDropdown(sid, data['submissions']);
                            FilesView.load(sid);
                        });
                    });
                });
            });

        });
    },

    subDropdown: function(id, vals) {
        for (var i = 0; i < vals.length; i++) {
            $('#submission-dropdown ul.dropdown-menu').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="#" submissionid="' + vals[i].Id + '">' + vals[i].User + ' at  ' + new Date(vals[i].Time).toLocaleString() + '</a></li>');
            if (id === vals[i].Id) {
                $('#submission-dropdown-label').attr('submissionid', id);
                $('#submission-dropdown-label').append('<h4><small>submission</small> ' + vals[i].User + ' at  ' + new Date(vals[i].Time).toLocaleString() + ' <span class="caret"></span></h4>');
            }
        }
        $('#submission-dropdown ul.dropdown-menu a').on('click', function() {
            $('#file-list').empty();
            var currentId = $(this).attr('submissionid');
            var currentName = $(this).html();
            var params = {
                'submission-id': currentId
            };
            setContext(params);
            $('#submission-dropdown-label').attr('submissionid', currentId);
            $('#submission-dropdown-label h4').html('<small>submission</small> ' + currentName + ' <span class="caret"></span>');
            FilesView.load(currentId);
        });

    },
    buidDropdown: function(tipe, id, url, vals) {
        for (var i = 0; i < vals.length; i++) {
            var currentId = tipe === 'user' ? vals[i].Name : vals[i].Id;
            var link = url + '?' + tipe + '-id=' + currentId;
            $('#' + tipe + '-dropdown ul.dropdown-menu').append('<li role="presentation"><a tabindex="-1" role="menuitem" href="' + link + '">' + vals[i].Name + '</a></li>');
            if (id === currentId) {
                $('#' + tipe + '-dropdown-label').attr(tipe + 'id', id);
                $('#' + tipe + '-dropdown-label').append('<h4><small>' + tipe + '</small> ' + vals[i].Name + ' <span class="caret"></span></h4>');
            }
        }
        if ($('#' + tipe + '-dropdown-label').attr(tipe + 'id') === undefined) {
            $('#' + tipe + '-dropdown-label').append('<h4><small>' + tipe + '</small> None Selected <span class="caret"></span></h4>');
        }
    },
    load: function(sid) {
        $.getJSON('fileinfos?submission-id=' + sid, function(data) {
            if (not(data['fileinfos'])) {
                return;
            }
            var fs = data['fileinfos'];
            for (var i = 0; i < fs.length; i++) {
                FilesView.addInfo(fs[i], sid);
            }
        });
    },
    addInfo: function(f, sid) {
        $.getJSON('resultnames?submission-id=' + sid + '&filename=' + f.Name, function(data) {
            var e = '<li><h3>' + f.Name + '</h3><h5>' + f.Count + ' Snapshots</h5>';
            var rs = data['resultnames'];
            if (not(rs)) {
                e += '<div class="alert alert-danger alert-dynamic alert-dismissable"><button type="button" class="close" data-dismiss="alert" aria-hidden="true">&times;</button><strong>No results available</strong></div>';
            } else {
                e += '<ul class="list-inline"><li class="dropdown"><button class="btn btn-primary dropdown-toggle" data-toggle="dropdown" href="#">Analysis <span class="caret"></span></button><ul class="dropdown-menu" role="menu">';
                for (var o in rs) {
                    if (not(rs[o])) {
                        e += '<li><a href="resultsview?result=' + o + '&file=' + f.Name + '">' + o + '</a></li>';
                    } else {
                        e += '<li class="dropdown-submenu"><a tabindex="-1" href="#">' + o + '</a><ul class="dropdown-menu" role="menu">';
                        for (var k in rs[o]) {
                            if (not(rs[o][k])) {
                                e += '<li><a href="resultsview?result=' + o + ':' + k + '&file=' + f.Name + '">' + k + '</a></li>';
                            } else {
                                e += '<li class="dropdown-submenu"><a tabindex="-1" href="#">' + k + '</a><ul class="dropdown-menu" role="menu">';
                                for (var j = 0; j < rs[o][k].length; j++) {
                                    e += '<li><a testid=' + rs[o][k][j] + ' href="resultsview?result=' + o + ':' + k + '-' + rs[o][k][j] + '&file=' + f.Name + '"></a></li>';
                                }
                                e += '</ul></li>';
                            }
                        }
                        e += '</ul></li>';
                    }
                }
                e += '</ul></li></ul>';
            }
            e += '</li>';
            $('#files-list').append(e);
            $('a[testid]').each(function() {
                var id = $(this).attr('testid');
                $.getJSON('files?id=' + id, function(data) {
                    if (not(data)) {
                        return;
                    }
                    $('a[testid="' + id + '"]').html(new Date(data['files'][0].Time).toLocaleString());
                });
            });
        });

    }
}