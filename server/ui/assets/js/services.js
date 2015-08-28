var anisyncServices = angular.module('anisyncServices', ['ngResource']);

anisyncServices.factory('Anisync', ['$resource',
  function($resource) {
    return {
      Check: $resource('api/check', {}, {
        query: {
          method: 'GET',
          params: {},
          isArray: true
        }
      }),
      Sync: $resource('api/sync', {}, {
        query: {
          method: 'POST',
          params: {},
          isArray: true
        }
      })
    }
  }
]);

anisyncServices.factory('Sync', ['$resource', ]);
