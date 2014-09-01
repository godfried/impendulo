var DeleteView = {
    init: function() {
        $(function() {
            $.getJSON('projects', function(data) {
                if (not(data['projects'])) {
                    return;
                }
                var ps = data['projects'];
                for (var i = 0; i < ps.length; i++) {
                    $('#tp-project-id').append('<option value="' + ps[i].Id + '">' + ps[i].Name + '</option>');
                    $('#ts-project-id').append('<option value="' + ps[i].Id + '">' + ps[i].Name + '</option>');
                    $('#tr-project-id').append('<option value="' + ps[i].Id + '">' + ps[i].Name + '</option>');
                    $('#tsk-project-id').append('<option value="' + ps[i].Id + '">' + ps[i].Name + '</option>');
                }
                $('#tp-project-id').multiselect({
                    selectedText: '# of # projects selected',
                    noneSelectedText: 'Delete projects'
                });
                $('#tp-project-id').multiselected = true;
                DeleteView.loadSubmissions('#ts-submission-id', ps[0].Id, 'Delete submissions');
                $('#ts-project-id').change(function() {
                    DeleteView.loadSubmissions('#ts-submission-id', $(this).val(), 'Delete submissions');
                });
                DeleteView.loadSubmissions('#tr-submission-id', ps[0].Id, 'Delete submission results');
                $('#tr-project-id').change(function() {
                    DeleteView.loadSubmissions('#tr-submission-id', $(this).val(), 'Delete submission results');
                });
                DeleteView.loadSkeletons(ps[0].Id);
                $('#tsk-project-id').change(function() {
                    DeleteView.loadSkeletons($(this).val());
                });
            });
            $.getJSON('usernames', function(data) {
                if (not(data['usernames'])) {
                    return;
                }
                var us = data['usernames'];
                for (var i = 0; i < us.length; i++) {
                    $('#tu-user-id').append('<option value="' + us[i] + '">' + us[i] + '</option>');
                }
                $('#tu-user-id').multiselect({
                    selectedText: '# of # users selected',
                    noneSelectedText: 'Delete users'
                });
                $('#tu-user-id').multiselected = true;
            });
        });
    },

    loadSkeletons: function(pid) {
        var id = '#tsk-skeleton-id';
        clearMulti(id);
        $(id).hide();
        $.getJSON('skeletons?project-id=' + pid, function(data) {
            if (not(data['skeletons'])) {
                return;
            }
            $(id).show();
            var sk = data['skeletons'];
            for (var i = 0; i < sk.length; i++) {
                $(id).append('<option value="' + sk[i].Id + '">' + sk[i].Name + '</option>');
            }
            $(id).multiselect({
                selectedText: '# of # skeletons selected',
                noneSelectedText: 'Delete skeletons'
            });
            $(id).multiselected = true;
        });
    },

    loadSubmissions: function(id, pid, desc) {
        clearMulti(id);
        $(id).hide();
        $.getJSON('submissions?project-id=' + pid, function(data) {
            if (not(data['submissions'])) {
                return;
            }
            $(id).show();
            var ss = data['submissions'];
            for (var i = 0; i < ss.length; i++) {
                DeleteView.loadSubmission(id, ss[i], desc, ss.length);
            }
        });
    },

    loadSubmission: function(id, sub, desc, num) {
        $.getJSON('counts?submission-id=' + sub.Id, function(data) {
            if (not(data['counts'])) {
                return;
            }
            var c = data['counts'];
            var l = c['launch'];
            var s = c['source'];
            var t = c['test'];
            $(id).append('<option date="' + new Date(sub.Time).toLocaleString() + '" source="' + s + '" launch="' + l + '" test="' + t + '" value="' + sub.Id + '">' + sub.User + '</option>');
            if ($(id).children().length === num) {
                $(id).multiselect({
                    selectedText: '# of # submissions selected',
                    noneSelectedText: desc,
                    classes: 'multiselect-submissions'
                });
                $('.multiselect-submissions .ui-multiselect-checkboxes li').tooltip({
                    title: function() {
                        var sl = 'option[value="' + $(this).find('input').val() + '"]';
                        var d = $(sl).attr('date');
                        var sc = $(sl).attr('source');
                        var lc = $(sl).attr('launch');
                        var tc = $(sl).attr('test');
                        return 'Date: ' + d + '\nSource Files: ' + sc + '\nLaunches: ' + lc + '\nTests: ' + tc;
                    },
                    placement: 'left',
                    container: 'body'
                });
                $(id).multiselected = true;
            }
        });
    }

}
