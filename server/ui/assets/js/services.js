var anisyncServices = angular.module('anisyncServices', ['ngResource']);

anisyncServices.factory('Anisync', ['$resource',
  function($resource) {
    return {
      Check: $resource('api/test/check', {}, {
        query: {
          method: 'GET',
          params: {}
        }
      }),
      Sync: $resource('api/sync', {}, {
        query: {
          method: 'POST',
          params: {}
        }
      })
    }
  }
]);

anisyncServices.factory('Sync', ['$resource', ]);
