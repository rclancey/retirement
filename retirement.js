var Retirement = function(path) {
  this.path = path;
};

Retirement.prototype.summary = function() {
  var self = this;
  return fetch(this.path+'/index.json', {
    method: 'get',
  }).then(function(response) {
    if (response.status == 200) {
      return response.json();
    }
    return response.text().then(function(text) {
      throw new Error(text);
    });
  }).then(function(data) {
    self.summaryData = data;
    self.drawBarCharts();
    return data;
  }).catch(function(err) {
    console.error("error loading summary: ", err);
  });
};

function clearNode(node) {
  while (node.firstChild) {
    node.removeChild(node.firstChild);
  }
}

function formatDollars(val) {
  var suf = '';
  if (Math.abs(val) >= 1000) {
    if (Math.abs(val) < 1000000) {
      val /= 1000;
      suf = 'k';
    } else {
      val /= 1000000;
      suf = 'M';
    }
  }
  if (val == 0) {
    return '$0';
  }
  if (val < 0) {
    return '-$' + val + suf;
  }
  return '$' + val + suf;
}

Retirement.prototype.applyFilters = function() {
  var self = this;
  var ageRange = [Number.POSITIVE_INFINITY, Number.NEGATIVE_INFINITY];
  var balRange = [Number.POSITIVE_INFINITY, Number.NEGATIVE_INFINITY];
  var mktRange = [Number.POSITIVE_INFINITY, Number.NEGATIVE_INFINITY];
  console.debug('ageRange = (%o, %o)', ageRange[0], ageRange[1]);
  var data = this.summaryData.runs.filter(function(run) {
    if (self.ageFilter) {
      if (run.death < self.ageFilter[0]) {
        return false;
      }
      if (run.death > self.ageFilter[1]) {
        return false;
      }
    }
    if (self.balanceFilter) {
      if (run.balance < self.balanceFilter[0]) {
        return false;
      }
      if (run.balance > self.balanceFilter[1]) {
        return false;
      }
    }
    if (self.marketFilter) {
      if (run.market * 100 < self.marketFilter[0]) {
        return false;
      }
      if (run.market * 100 > self.marketFilter[1]) {
        return false;
      }
    }
    if (run.death < ageRange[0]) {
      ageRange[0] = run.death;
    }
    if (run.death > ageRange[1]) {
      ageRange[1] = run.death;
    }
    if (run.balance < balRange[0]) {
      balRange[0] = run.balance;
    }
    if (run.balance > balRange[1]) {
      balRange[1] = run.balance;
    }
    if (run.market * 100 < mktRange[0]) {
      mktRange[0] = run.market * 100;
    }
    if (run.market * 100 > mktRange[1]) {
      mktRange[1] = run.market * 100;
    }
    return true;
  });
  console.debug('ageRange = (%o, %o)', ageRange[0], ageRange[1]);
  ageRange[0] = 2 * Math.floor(ageRange[0] / 2);
  ageRange[1] = 2 * Math.ceil(ageRange[1] / 2);
  balRange[0] = 250000 * Math.floor(balRange[0] / 250000);
  balRange[1] = 250000 * Math.ceil(balRange[1] / 250000);
  mktRange[0] = 0.5 * Math.floor(mktRange[0] / 0.5);
  mktRange[1] = 0.5 * Math.ceil(mktRange[1] / 0.5);

  var div = document.querySelector('#age .filter');
  clearNode(div);
  div.appendChild(document.createTextNode("Age Range: " + ageRange[0] + " - " + ageRange[1]));
  if (this.ageFilter) {
    var clear = document.createElement('span');
    clear.appendChild(document.createTextNode('Clear Filter'));
    clear.className = 'clearFilter';
    clear.addEventListener('click', this.clearAgeFilter.bind(this));
    div.appendChild(clear);
  }

  div = document.querySelector('#balance .filter');
  clearNode(div);
  div.appendChild(document.createTextNode("Balance Range: " + formatDollars(balRange[0]) + " - " + formatDollars(balRange[1])));
  if (this.balanceFilter) {
    var clear = document.createElement('span');
    clear.appendChild(document.createTextNode('Clear Filter'));
    clear.className = 'clearFilter';
    clear.addEventListener('click', this.clearBalanceFilter.bind(this));
    div.appendChild(clear);
  }

  div = document.querySelector('#market .filter');
  clearNode(div);
  div.appendChild(document.createTextNode("Market Range: " + mktRange[0] + " - " + mktRange[1]));
  if (this.marketFilter) {
    var clear = document.createElement('span');
    clear.appendChild(document.createTextNode('Clear Filter'));
    clear.className = 'clearFilter';
    clear.addEventListener('click', this.clearMarketFilter.bind(this));
    div.appendChild(clear);
  }

  return data;
};

Retirement.prototype.drawBarCharts = function() {
  var self = this;
  var data = this.applyFilters();
  var age = {
    type: 'histogram',
    x: data.map(function(run) { return run.death }),
    autobinx: false,
    xbins: {
      start: 40,
      end: 100,
      size: 2,
    },
  };
  var balance = {
    type: 'histogram',
    x: data.map(function(run) { return run.balance }),
    autobinx: false,
    xbins: {
      start: -2000000,
      end: 10000000,
      size: 250000,
    },
  };
  var market = {
    type: 'histogram',
    x: data.map(function(run) { return run.market * 100 }),
    autobinx: false,
    xbins: {
      start: -10,
      end: 12,
      size: 0.5,
    },
  };
  var layout = {
    margin: { t: 10, b: 25, r: 15, l: 25 },
    bargap: 0.1,
    bargroupgap: 0.2,
  };
  var config = {
    displayModeBar: false,
  };
  var ageCanvas = document.querySelector('#age .chart');
  var balCanvas = document.querySelector('#balance .chart');
  var mktCanvas = document.querySelector('#market .chart');
  Plotly.newPlot(ageCanvas, [age], layout, config);
  Plotly.newPlot(balCanvas, [balance], layout, config);
  Plotly.newPlot(mktCanvas, [market], layout, config);
  var barClicker = function(evt) {
    var node = evt.event.target;
    if (node.localName == 'rect') {
      var bottom = node.y.baseVal.value;
      var height = node.height.baseVal.value;
      var y = evt.event.layerY;
      var by = height + bottom - y;
      var dh = evt.points[0].yaxis._tmax - evt.points[0].yaxis._tmin;
      var v = Math.round((by / height) * dh);
      if (v < 0) {
        v = 0;
      }
      if (v >= evt.points[0].pointIndices.length) {
        v = evt.points[0].pointIndices.length - 1;
      }
      var idx = evt.points[0].pointIndices[v];
      idx = data[idx].index;
      console.debug('bottom=%o, h=%o, y=%o, dh=%o, v=%o, idx=%o', bottom, height, y, dh, v, idx);
      self.balances(idx);
    }
  };
  ageCanvas.on('plotly_click', barClicker);
  balCanvas.on('plotly_click', barClicker);
  mktCanvas.on('plotly_click', barClicker);
  ageCanvas.on('plotly_relayout', function(evt) {
    if (evt['xaxis.autorange']) {
      self.clearAgeFilter();
    } else {
      self.setAgeFilter(evt['xaxis.range[0]'], evt['xaxis.range[1]']);
    }
  });
  balCanvas.on('plotly_relayout', function(evt) {
    if (evt['xaxis.autorange']) {
      self.clearBalanceFilter();
    } else {
      self.setBalanceFilter(evt['xaxis.range[0]'], evt['xaxis.range[1]']);
    }
  });
  mktCanvas.on('plotly_relayout', function(evt) {
    if (evt['xaxis.autorange']) {
      self.clearBalanceFilter();
    } else {
      self.setMarketFilter(evt['xaxis.range[0]'], evt['xaxis.range[1]']);
    }
  });
};

Retirement.prototype.clearAgeFilter = function() {
  this.ageFilter = null;
  this.drawBarCharts();
};

Retirement.prototype.setAgeFilter = function(a, b) {
  this.ageFilter = [a, b];
  this.drawBarCharts();
};

Retirement.prototype.clearBalanceFilter = function() {
  this.balanceFilter = null;
  this.drawBarCharts();
};

Retirement.prototype.setBalanceFilter = function(a, b) {
  this.balanceFilter = [a, b];
  this.drawBarCharts();
};

Retirement.prototype.clearMarketFilter = function() {
  this.marketFilter = null;
  this.drawBarCharts();
};

Retirement.prototype.setMarketFilter = function(a, b) {
  this.marketFilter = [a, b];
  this.drawBarCharts();
};

var evStyles = [
    { bgcolor: 'rgba(255,255,255,0.5)' },
    { bgcolor: 'rgba(255,255,255,0.5)' },
    { bgcolor: 'rgba(255,255,255,0.5)' },
    { bgcolor: 'rgba(127,255,127,0.5)' },
    { bgcolor: 'rgba(127,255,127,0.5)' },
    { bgcolor: 'rgba(127,127,255,0.5)' },
    { bgcolor: 'rgba(127,127,255,0.5)' },
    { bgcolor: 'rgba(255,0,0,0.5)', font: { color: 'white' } },
    { bgcolor: 'rgba(255,0,0,0.75)', font: { color: 'white' } },
    { bgcolor: 'rgba(255,0,0,0.75)', font: { color: 'white' } },
    { bgcolor: 'rgba(255,0,0,1.0)', font: { color: 'white' } },
];

var colors = [

  'rgb(127, 0, 0)',
  'rgb(0, 127, 0)',
  'rgb(0, 0, 127)',

  'rgb(127, 127, 0)',
  'rgb(127, 0, 127)',
  'rgb(0, 127, 127)',

  'rgb(255, 127, 0)',
  'rgb(255, 0, 127)',
  'rgb(127, 255, 0)',
  'rgb(127, 0, 255)',
  'rgb(0, 255, 127)',
  'rgb(0, 127, 255)',

  'rgb(255, 127, 127)',
  'rgb(127, 255, 127)',
  'rgb(127, 127, 255)',
  'rgb(255, 255, 127)',
  'rgb(255, 127, 255)',
  'rgb(127, 255, 255)',

];

Retirement.prototype.balances = function(i) {
  var n = ''+i;
  while (n.length < 6) {
    n = '0' + n;
  }
  Plotly.d3.csv('go/results/balances-'+n+'.csv', function(balances) {
    var first = Object.assign({}, balances[0]);
    delete(first.date);
    var series = Object.keys(first);
    series.sort();
    var data = series.map(function(k, i) {
      return {
        type: 'scatter',
        mode: 'lines',
        line: {
          color: k == 'Total' ? 'rgb(0, 0, 0)' : colors[i],
          width: k == 'Total' ? 3 : 2,
        },
        name: k,
        x: balances.map(function(row) { return row.date }),
        y: balances.map(function(row) { return row[k] ? parseFloat(row[k]) : null }),
      }
    });
    Plotly.d3.csv('go/results/events-'+n+'.csv', function(evdata) {
      var events = evdata.filter(function(ev) {
        ev.severity = parseInt(ev.severity);
        return ev.severity > 5;
      }).map(function(ev, i) {
        return Object.assign({
          x: ev.date,
          y: 0,
          xref: 'x',
          yref: 'y',
          xanchor: 'right',
          text: ev.value,
          bgcolor: 'rgba(255,255,255,0.5)',
          arrowwidth: 0.5,
          arrowcolor: 'rgba(0,0,0,0.5)',
          showarrow: true,
          arrowhead: 7,
          ax: 0,
          ay: -20 * ((i % 10) + 2),
        }, evStyles[ev.severity-1]);
      });
      var layout = {
        showlegend: false,
        margin: { t: 0 },
        annotations: events,
      };
      Plotly.newPlot(
        document.getElementById('chart'),
        data,
        layout
      );
    });
  });
};

