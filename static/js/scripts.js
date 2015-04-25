(function() {
    var app = angular.module('scripts', []);

    plots = [];
    plotsData = {};
    plotsStatus = {};
    plotsLastValue = {};

    app.checkStatus = function(name) {
        if (typeof(plotsLastValue[name]) != "undefined" && typeof(plotsStatus[name]) != "undefined") {
            var lastValue = plotsLastValue[name][1];
            var statusSort = plotsStatus[name].status_sort;
            var status = plotsStatus[name].status;

            var statusClass;

            if (statusSort == "desc") {
                if (lastValue < status.ok) {
                    statusClass = "success";
                } else if (lastValue < status.warning) {
                    statusClass = "warning";
                } else {
                    statusClass = "danger";
                }
            } else {
                if (lastValue < status.critical) {
                    statusClass = "danger";
                } else if (lastValue < status.warning) {
                    statusClass = "warning";
                } else {
                    statusClass = "success";
                }
            }

            $('#'+name+'status').removeClass(function (index, css) {
                return (css.match (/(^|\s)label-\S+/g) || []).join(' ');
            });

            $('#'+name+'status').addClass("label-"+statusClass);
            $('#'+name+'status').text(statusClass.toUpperCase());

            $('#'+name+'lastvalue').text("Last Value: "+lastValue);
        }
    };

    app.run(['$location', function($location) {
        var conn = new WebSocket("ws://"+$location.host()+":"+$location.port()+"/ws");
        conn.onclose = function(evt) {
            $('#errormodal').modal('show');
        };
        conn.onmessage = function(evt) {
            var result = $.parseJSON(evt.data)[1].split(':');
            plotsData[result[0]] = plotsData[result[0]].slice(1);
            plotsData[result[0]].push([result[1] * 1000, result[2]]);
            plots[result[0]].setData([plotsData[result[0]]]);
            plots[result[0]].setupGrid();
            plots[result[0]].draw();

            plotsLastValue[result[0]] = [result[1] * 1000, result[2]];

            app.checkStatus(result[0]);
        };
    }]);


    timezoneJS.timezone.zoneFileBasePath = "static/tz";
    timezoneJS.timezone.defaultZoneFile = [];
    timezoneJS.timezone.init({async: false});

    app.directive('scriptInfoModal', function() {
        return {
            restrict: 'E',
            scope: {
                name: '@name'
            },
            templateUrl: 'static/tmpl/script-info-modal.html?v=0.1',
            controller: ['$http', '$scope', function($http, $scope) {
                $scope.last_value = null;

                $http.get('/scripts/'+$scope.name).success(function(data) {
                    $scope.details = data;

                    plotsStatus[$scope.name] = {status: data.status, status_sort: data.status_sort};

                    $scope.last_value = plotsLastValue[$scope.name][1];

                    app.checkStatus($scope.name);
                });
            }],
            controllerAs: 'script',
        };
    });


    app.directive('scriptInfo', function() {
        return {
            restrict: 'E',
            scope: {
                name: '@name'
            },
            templateUrl: 'static/tmpl/script-info.html?v=0.1',
        };
    });

    app.directive('scriptChart', function() {
        return {
            restrict: 'E',
            scope: {
                name: '@name'
            },
            templateUrl: 'static/tmpl/script-chart.html?v=0.1',
            controller: ['$http', '$scope', function($http, $scope) {
                this.period = 'hour';

                var chart = this;

                this.draw = function() {
                    $http.get('/scripts/chart/'+$scope.name+'?period='+chart.period).success(function(data) {
                        var chartData = [];

                        $.each(data, function(index, key) {
                            var value = index.split(':')[1];

                            chartData.push([key * 1000, value]);
                        });

                        plotsData[$scope.name] = chartData;

                        plots[$scope.name] = $.plot("#chart"+$scope.name, [chartData], {series:{shadowSize:0}, xaxis: {mode: "time", timezone:"browser"}});

                        plotsLastValue[$scope.name] = chartData[chartData.length-1];
                    });
                };

                this.draw();
            }],
            controllerAs: 'script',
        };
    });
})();
