function overviewChart(chartData, tipe){
    if (chartData === null){
	return;
    }
    var m = [10, 150, 100, 100];
    var w = 1100 - m[1] - m[3];
    var h = 500 - m[0] - m[2];
    var size = w / chartData.length;
    var yMax = h/3;
    var duration = 1500;
    var ease = 'quad-in-out';
    var names = chartData.map(function(d){return d.key;});
    var categories = ['Submissions', 'Snapshots', 'Launches'];
    var chartColour = function(tipe) { 
	return d3.scale.category10()
            .domain(categories)(tipe); 
    };    

    var xScale = d3.scale.ordinal()
	.domain(names)
	.rangeBands([0, w]);
    
    var xPos = function(d){
	return xScale(d.key);
    };
    var subDomain = [0, d3.max(chartData, submissions)];
    var snapDomain = [0, d3.max(chartData, snapshots)];
    var launchDomain = [0, d3.max(chartData, launches)];
   
    var subPos = function(d){
	return h - subHeight(d);
    };
    var snapPos = function(d){
	return subPos(d) - snapHeight(d);
    };

    var launchPos = function(d){
	return snapPos(d) - launchHeight(d);
    };

    var subHeight = function(d){
	var ret = d3.scale.linear()
	    .domain(subDomain)
	    .range([0, yMax])(d.submissions);
	return ret;
    };
    var snapHeight = function(d){
	var ret = d3.scale.linear()
	    .domain(snapDomain)
	    .range([0, yMax])(d.snapshots);
	return ret;
    };

    var launchHeight = function(d){
	var ret = d3.scale.linear()
	    .domain(launchDomain)
	    .range([0, yMax])(d.launches);
	return ret;
    };
    var groupBuffer = 5;
    var x = w/chartData.length - 2*groupBuffer;

    var groupWidth = x/3;
        
    var groupSub = function(d){
	var ret = d3.scale.linear()
	    .domain(subDomain)
	    .range([0, h])(d.submissions);
	return ret;
    };
    var groupSnap = function(d){
	var ret = d3.scale.linear()
	    .domain(snapDomain)
	    .range([0, h])(d.snapshots);
	return ret;
    };

    var groupLaunch = function(d){
	var ret = d3.scale.linear()
	    .domain(launchDomain)
	    .range([0, h])(d.launches);
	return ret;
    };

    var transitionStacked = function(){
	bars.selectAll('.submission')
	    .transition()
	    .duration(duration)
	    .ease(ease)
	    .attr('status', 'stacked')
	    .attr("x", xPos)
	    .attr("y", subPos)
	    .attr("width", x)
	    .attr("height", subHeight);
	
	bars.selectAll('.snapshot')
	    .transition()
	    .duration(duration)
	    .ease(ease)
	    .attr("x", xPos)
	    .attr("y", snapPos)
	    .attr("width", x)
	    .attr("height", snapHeight);

	bars.selectAll('.launch')
	    .transition()
	    .duration(duration)
	    .ease(ease)
	    .attr("x", xPos)
	    .attr("y", launchPos)
	    .attr("width", x)
	    .attr("height", launchHeight);
	
	bars.selectAll('.subinfo')
	    .transition()
	    .duration(duration)
	    .ease(ease)
	    .attr("x", function(d) { 
		return xPos(d) + x/2.5; 
	    })
	    .attr("y", function(d) {
		return h-2;
	    });

	bars.selectAll('.snapinfo')
	    .transition()
	    .duration(duration)
	    .ease(ease)
	    .attr("x", function(d) { 
		return xPos(d) + x/2.5; 
	    })
	    .attr("y", function(d) {
		return subPos(d)-2;
	    });

	bars.selectAll('.launchinfo')
	    .transition()
	    .duration(duration)
	    .ease(ease)
	    .attr("x", function(d) { 
		return xPos(d) + x/2.5; 
	    })
	    .attr("y", function(d) {
		return snapPos(d)-2;
	    });

    };

    var transitionGrouped = function(){
	bars.selectAll('.submission')
	    .transition()
	    .duration(duration)
	    .ease(ease)
	    .attr('status', 'grouped')
	    .attr("x", function(d) { 
		return xPos(d)+groupBuffer; 
	    })
	    .attr("y", function(d){return h - groupSub(d);})
	    .attr("width", groupWidth)
	    .attr("height", groupSub);
	
	bars.selectAll('.snapshot')
	    .transition()
	    .duration(duration)
	    .ease(ease)
	    .attr("x", function(d) { 
		return xPos(d)+groupWidth + groupBuffer; 
	    })
	    .attr("y", function(d){return h - groupSnap(d);})
	    .attr("width", groupWidth)
	    .attr("height", groupSnap);

	bars.selectAll('.launch')
	    .transition()
	    .duration(duration)
	    .ease(ease)
	    .attr("x", function(d) { 
		return xPos(d)+groupWidth*2 + groupBuffer; 
	    })
	    .attr("y", function(d){return h - groupLaunch(d);})
	    .attr("width", groupWidth)
	    .attr("height", groupLaunch);
	
	bars.selectAll('.subinfo')
	    .transition()
	    .duration(duration)
	    .ease(ease)
	    .attr("x", function(d) { 
		return xPos(d)+groupWidth * 0.2 + groupBuffer; 
	    })
	    .attr("y", function(d){return h - groupSub(d) - 2;});

	bars.selectAll('.snapinfo')
	    .transition()
	    .duration(duration)
	    .ease(ease)
	    .attr("x", function(d) { 
		return xPos(d)+groupWidth * 1.2 + groupBuffer; 
	    })
	    .attr("y", function(d){return h - groupSnap(d) - 2;});

	bars.selectAll('.launchinfo')
	    .transition()
	    .duration(duration)
	    .ease(ease)
	    .attr("x", function(d) {
		return xPos(d) + groupWidth * 2.2 + groupBuffer; 
	    })
	    .attr("y", function(d){return h - groupLaunch(d) - 2;});
    };

    var change = function(d) {
	var status = d3.select('.submission')
	    .attr('status');
	if (status === 'stacked'){
	    transitionGrouped();
	} else{
	    transitionStacked();
	}
    };
    
    var xAxis = d3.svg.axis()
	.scale(xScale)
	.tickSize(-h)
	.orient('bottom')
	.tickSubdivide(true);

    var chart = d3.select('#chart')
	.append('svg:svg')
	.attr('width', w + m[1] + m[3])
	.attr('height', h + m[0] + m[2])
    	.append('svg:g')
	.attr('transform', 'translate(' + m[3] + ',' + m[0] + ')');

    chart.append('svg:rect')
	.attr('width', w)
	.attr('height', h)
	.attr('class', 'plot');
    
    chart.append('svg:g')
	.attr('font-size','10px')
	.attr('class', 'x axis')
	.attr('transform', 'translate(0,' + h + ')')
	.call(xAxis);

    chart.append('text')
        .attr('x', w/2 )
        .attr('y',  h+40)
	.attr('font-size','20px')
        .style('text-anchor', 'middle')
        .text(tipe === 'project' ? 'Project' : 'User');
    
    chart.append('svg:clipPath')
	.attr('id', 'clip')
	.append('svg:rect')
	.attr('x', -10)
	.attr('y', -10)
	.attr('width', w+20)
	.attr('height', h+20);

    var chartBody = chart.append('g')
	.attr('clip-path', 'url(#clip)')
	.attr('class', 'clickable')
	.on('click', change);

    var bars = chartBody.selectAll('.bars')
	.data(chartData)
	.enter()
	.append('g')
	.attr('class', 'bar');
    
    bars.append('rect')
	.attr('class', 'submission')
	.attr("x", xPos)
	.attr('fill', chartColour('Submissions'))
	.attr("y", subPos)
	.attr("width", x)
	.attr("height", subHeight)
	.attr('status', 'stacked')
	.append('title')
	.attr('class', 'description')
	.text(function(d) { 
	    return d.key + 
		'\n'+d.submissions+' submissions';
	});

    bars.append("rect")
	.attr('class', 'snapshot')
	.attr("x", xPos)
	.attr('fill', chartColour('Snapshots'))
	.attr("y", snapPos)
	.attr("width", x)
	.attr("height", snapHeight)
	.append('title')
	.attr('class', 'description')
	.text(function(d) { 
	    return d.key + 
		'\n'+d.snapshots+' snapshots';
	});
  
    bars.append("rect")
	.attr('class', 'launch')
	.attr("x", xPos)
	.attr('fill', chartColour('Launches'))
	.attr("y", launchPos)
	.attr("width", x)
	.attr("height", launchHeight)
	.append('title')
	.attr('class', 'description')
	.text(function(d) { 
	    return d.key + 
		'\n'+d.launches+' launches';
	});

    bars.append('text')
	.attr('class', 'subinfo')
	.attr("x", function(d) { 
	    return xPos(d) + x/2.5; 
	})
	.attr("y", function(d) {
	    return h-2;
	})
	.attr('font-size', '9px')
	.text(submissions);

    bars.append('text')
	.attr('class', 'snapinfo')
	.attr("x", function(d) { 
	    return xPos(d) + x/2.5; 
	})
	.attr("y", function(d) {
	    return subPos(d)-2;
	})
	.attr('font-size', '9px')
	.text(snapshots);

    bars.append('text')
	.attr('class', 'launchinfo')
	.attr("x", function(d) { 
	    return xPos(d) + x/2.5; 
	})
	.attr("y", function(d) {
	    return snapPos(d)-2;
	})
	.attr('font-size', '9px')
	.text(launches);

    var legend = chart.append('g')
	.attr('class', 'legend')
    	.attr('height', 100)
	.attr('width', 100)
	.attr('transform', 'translate(-100,0)');  
    
    var legendElements = legend.selectAll('g')
	.data(categories)
	.enter()
	.append('g');

    legendElements.append('text')
	.attr('x', 20)
	.attr('y', function(d, i){
	    return i*20+60;
	})
	.attr('font-size','12px')
	.text(function(d){
	    return d;
	});
    
    legendElements.append('rect')
	.attr('class', 'legendrect')
	.attr('x', 0)
	.attr('y', function(d, i){ 
	    return i*20 + 50;
	})
	.attr('width', 15)
	.attr('height', 15)
	.style('fill', chartColour);
  
}

function submissions(d){
    return d.submissions;
}


function snapshots(d){
    return d.snapshots;
}


function launches(d){
    return d.launches;
}
