//Copyright (c) 2013, The Impendulo Authors
//All rights reserved.
//
//Redistribution and use in source and binary forms, with or without modification,
//are permitted provided that the following conditions are met:
//
//  Redistributions of source code must retain the above copyright notice, this
//  list of conditions and the following disclaimer.
//
//  Redistributions in binary form must reproduce the above copyright notice, this
//  list of conditions and the following disclaimer in the documentation and/or
//  other materials provided with the distribution.
//
//THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
//ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
//WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
//DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
//ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
//(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
//LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
//ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
//(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

function movevalue(srcParentId, destParentId, src) {
    var dest = document.createElement('option');
    dest.innerHTML = src.innerHTML;
    dest.setAttribute('value', src.value);
    dest.setAttribute('onclick', "movevalue('"+destParentId+"', '"+srcParentId+"', this)");
    var destParent = document.getElementById(destParentId);
    var srcParent = document.getElementById(srcParentId);
    srcParent.removeChild(src);
    if(destParent.getAttribute('added') === 'true'){
	dest.setAttribute('selected', true);
    } else{
	var nodes = srcParent.childNodes;
	for(var i=0; i<nodes.length; i++) {
	    if (nodes[i].nodeName.toLowerCase() == 'option') {
		nodes[i].setAttribute('selected', true);
	    }
	}
    }
    destParent.appendChild(dest);
}

function unhide(it, box) {
    var check = (box.checked) ? 'block' : 'none';
    document.getElementById(it).style.display = check;
}

function replacevalue(srcParentID, destID, src) {
    var dest = document.getElementById(destID);
    if(dest === null){
	return;
    }
    var srcParent = document.getElementById(srcParentID);
    if(dest.value != ''){
	var newChild = document.createElement('option');
	newChild.innerHTML = dest.value;
	newChild.setAttribute('value', dest.value);
	newChild.setAttribute('onclick', "replacevalue('"+srcParentID+"', '"+destID+"', this)");
    	srcParent.appendChild(newChild);
    }
    dest.setAttribute('value', src.value);
    dest.setAttribute('onclick', "movevalueback('"+srcParentID+"', this)");
    srcParent.removeChild(src);
}

function movevalueback(destParent, src) {
    var dest = document.createElement('option');
    dest.innerHTML = src.value;
    dest.setAttribute('value', src.value);
    dest.setAttribute('onclick', "replacevalue('"+destParent+"', '"+src.getAttribute("id")+"', this)");
    document.getElementById(destParent).appendChild(dest);
    src.setAttribute('value', '');
}

function showdescription(description) {
    document.getElementById('description').innerHTML = description;
}

function movedescriptionvalue(srcParent, destParent, srcId) {
    var src = document.getElementById(srcId);
    var id = src.getAttribute('ruleid');
    var name = src.getAttribute('rulename');
    var description = src.getAttribute('ruledescription');
    var dest = document.createElement('option');
    dest.innerHTML = name;
    dest.setAttribute('ruleid', id);
    dest.setAttribute('rulename', name);
    dest.setAttribute('ruledescription', description);
    dest.setAttribute('onclick', "showdescription('"+description+"')");
    dest.setAttribute('ondblclick', "addalert('"+destParent+"', '"+srcParent+"', this)");
    document.getElementById(destParent).appendChild(dest);
    document.getElementById(srcParent).removeChild(src);
}

function addalert(srcParent, destParent, src) {
    var id = src.getAttribute('ruleid');
    var name = src.getAttribute('rulename');
    var description = src.getAttribute('ruledescription');
    var dest = document.createElement('div');
    dest.setAttribute("class", 'alert alert-dismissable alert-list');
    dest.setAttribute('id', id);
    dest.setAttribute('ruleid', id);
    dest.setAttribute('rulename', name);
    dest.setAttribute('ruledescription', description);
    var destButton = document.createElement('button');
    destButton.setAttribute('class', 'close');
    destButton.setAttribute('type', 'button');
    destButton.setAttribute('data-dismiss', 'alert');
    destButton.setAttribute('aria-hidden', 'true');
    destButton.setAttribute('onclick', "movedescriptionvalue('"+destParent+"','"+srcParent+"', '"+id+"')");
    destButton.innerHTML = '&times;';
    var destName = document.createElement('strong');
    destName.innerHTML = name+': ';
    var destDescription = document.createElement('small');
    destDescription.setAttribute('class', 'text-muted');
    destDescription.innerHTML = description;
    var destAnchor = document.createElement('input');
    destAnchor.setAttribute('type', 'hidden');
    destAnchor.setAttribute('name', 'ruleid');
    destAnchor.setAttribute('value', id);
    dest.appendChild(destButton);
    dest.appendChild(destName);
    dest.appendChild(destDescription);
    dest.appendChild(destAnchor);
    document.getElementById(destParent).appendChild(dest);
    document.getElementById(srcParent).removeChild(src);
}

function highlight(){
    SyntaxHighlighter.defaults['toolbar'] = false;
    SyntaxHighlighter.defaults['class-name'] = 'error';
    SyntaxHighlighter.all();
}

function addSkeletons(src, dest, skeletonMap){
    var srcList = document.getElementById(src);
    var id = srcList.options[srcList.selectedIndex].value;
    $.getJSON('skeletons?projectid='+id, function(data){   
	var destList = document.getElementById(dest);
	destList.options.length = 0;
	if (data.skeletons === null || data.skeletons.length === 0){
	    return;
	}
	for(var i = 0; i < data.skeletons.length; i++) {
	    var option = document.createElement('option');
	    option.value = data.skeletons[i].Id;
	    option.text = data.skeletons[i].Name;
	    destList.add(option);
	}
    });
}

function populate(src, toolDest, userDest){
    addTools(src, toolDest);
    addUsers(src, userDest);
}

function ajaxSelect(src, dest, url, name){
    var srcList = document.getElementById(src);
    var val = srcList.options[srcList.selectedIndex].value;
    $.getJSON(url+val, function(data){   
	    $('#'+dest).multiselect();
	    $('#'+dest).multiselect('destroy');
	    var destList = document.getElementById(dest);
	    destList.options.length = 0;
	    var items = data[name];
	    for(var i = 0; i < items.length; i++) {
		var option = document.createElement('option');
		option.value = option.text = items[i];
		destList.add(option);
	    }
	    $('#'+dest).multiselect();
	    $('#'+dest).multiselected = true;
    });	 
}

function addTools(src, dest){
    ajaxSelect(src, dest, 'tools?projectid=', 'tools');
}
 
function addUsers(src, dest){
   ajaxSelect(src, dest, 'users?projectid=', 'users');
}


function addPopover(dest, src){
    $('body').on('click', function (e) {
        if ($('#'+dest).next('div.popover:visible').length > 0
	    && $(e.target).data('toggle') !== 'popover'
            && e.target.id !== dest
	    && e.target.id !== 'codepopover'
	    && $('#codepopover').find($(e.target)).length === 0){
  		$('#'+dest).click();
        }    
    });
    window.onload = function() {
	$('#'+dest).popover({
	    template: '<div id="codepopover" class="popover code-popover"><div class="arrow"></div><div class="popover-inner"><h3 class="popover-title"></h3><div class="popover-content code-popover-content"><p></p></div></div></div>',
	    placement : 'bottom', 
	    html: 'true',
	    content :  $('#'+src).html(),
	});
    };
}


function addCodeModal(dest, resultId, bug, start, end){
    $('#'+dest).click(function(){
	var id = dest+'modal';
	var s = '#'+id;
	if($(s).length > 0){
	    $(s).modal('show');
	    $(s).on('shown.bs.modal', function(e){
		line.scrollIntoView(); 
	    });
	    return;
	}
	$.getJSON('code?resultid='+resultId, function(data){
	    var h = 'highlight: [';
	    for(var i = start; i < end; i ++){
		h += i + ',';
	    }
	    h = h + end + '];'
	    var preClass = "'brush: java; "+h+"'";
	    jQuery("<div id="+id+" class='modal fade' tabindex='-1' role='dialog' aria-labelledby='"+id+"label' aria-hidden='true'><div class='modal-dialog'><div class='modal-content'><div class='modal-header'><button type='button' class='close' data-dismiss='modal' aria-hidden='true'>&times;</button><h4 class='modal-title' id='"+id+"label'>"+bug.title+" <br><small>"+bug.content+"</small></h4></div><div class='modal-body'><pre class="+preClass+">"+data.code+"</pre></div></div></div></div>").appendTo('body');
	    SyntaxHighlighter.defaults['toolbar'] = false;
	    SyntaxHighlighter.defaults['class-name'] = 'error';
	    SyntaxHighlighter.highlight(); 
	    $(s).find('.highlighted').attr('style', 'background-color: #ff7777 !important;');
	    $(s).modal('show');
	    $(s).on('shown.bs.modal', function(e){
		var offset = $(s).find('.highlighted').offset();
		var offsetParent = $(s).offset();
		$(s).animate({
		    scrollTop: offset.top - offsetParent.top
		});
	    });
	});
    });
}

function ajaxChart(subID, file, result, currentTime, nextTime, childID){
    var url = 'chart?subid='+subID+'&file='+file+'&result='+result;
    if(childID !== undefined){
	url += '&childfileid='+childID;
    }
    if(currentTime === undefined){
	currentTime = -1;
    }
    if(nextTime === undefined){
	nextTime = -1;
    }
    $.getJSON(url, function(data){
	showChart(name, name, data, false, currentTime, nextTime);
    });
}
