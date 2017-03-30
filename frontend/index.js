let map;
let data;

function geoJsonFromResponse(response) {
  const geojson = {
    type: 'FeatureCollection',
    features: [],
  };

  for (let i = 0; i < response.length; i++) {
    const position = response[i];
    const lat = position.Lat
    const lng = position.Lng
    const feature = {
      type: 'Feature',
      geometry: {
	type: 'Point',
	coordinates: [lng, lat],
      },
      properties: {
	mmsi: position.MMSI,
      },
    };

    geojson.features.push(feature);
  }

  return geojson;
}

function ajaxPromise(path) {
  const request = new Request(path, {
  });
  return fetch(request).then(response => {
    if (!response.ok) {
      // Note: this assumes that bad responses still return JSON data.
      return response.json().then(json => Promise.reject(json));
    }
    return response.json();
  });
}

function fetchPositions() {
  const endpoint = `http://olmsted.sidewalk.local:8000/positions`
  ajaxPromise(endpoint).then(response => {
    // Clear the old data before creating the new one.
    if (data) {
      data.setMap(null);
    }

    data = new google.maps.Data();
    data.addGeoJson(geoJsonFromResponse(response));
    data.setMap(map);
  });
}


function initMap() {
  map = new google.maps.Map(document.getElementById('map'), {
    styles: [
      {
	"elementType": "geometry",
	"stylers": [
	  {
	    "color": "#f5f5f5"
	  }
	]
      },
      {
	"elementType": "labels.icon",
	"stylers": [
	  {
	    "visibility": "off"
	  }
	]
      },
      {
	"elementType": "labels.text.fill",
	"stylers": [
	  {
	    "color": "#616161"
	  }
	]
      },
      {
	"elementType": "labels.text.stroke",
	"stylers": [
	  {
	    "color": "#f5f5f5"
	  }
	]
      },
      {
	"featureType": "administrative.land_parcel",
	"elementType": "labels.text.fill",
	"stylers": [
	  {
	    "color": "#bdbdbd"
	  }
	]
      },
      {
	"featureType": "poi",
	"elementType": "geometry",
	"stylers": [
	  {
	    "color": "#eeeeee"
	  }
	]
      },
      {
	"featureType": "poi",
	"elementType": "labels.text.fill",
	"stylers": [
	  {
	    "color": "#757575"
	  }
	]
      },
      {
	"featureType": "poi.park",
	"elementType": "geometry",
	"stylers": [
	  {
	    "color": "#e5e5e5"
	  }
	]
      },
      {
	"featureType": "poi.park",
	"elementType": "labels.text.fill",
	"stylers": [
	  {
	    "color": "#9e9e9e"
	  }
	]
      },
      {
	"featureType": "road",
	"elementType": "geometry",
	"stylers": [
	  {
	    "color": "#ffffff"
	  }
	]
      },
      {
	"featureType": "road.arterial",
	"elementType": "labels.text.fill",
	"stylers": [
	  {
	    "color": "#757575"
	  }
	]
      },
      {
	"featureType": "road.highway",
	"elementType": "geometry",
	"stylers": [
	  {
	    "color": "#dadada"
	  }
	]
      },
      {
	"featureType": "road.highway",
	"elementType": "labels.text.fill",
	"stylers": [
	  {
	    "color": "#616161"
	  }
	]
      },
      {
	"featureType": "road.local",
	"elementType": "labels.text.fill",
	"stylers": [
	  {
	    "color": "#9e9e9e"
	  }
	]
      },
      {
	"featureType": "transit.line",
	"elementType": "geometry",
	"stylers": [
	  {
	    "color": "#e5e5e5"
	  }
	]
      },
      {
	"featureType": "transit.station",
	"elementType": "geometry",
	"stylers": [
	  {
	    "color": "#eeeeee"
	  }
	]
      },
      {
	"featureType": "water",
	"elementType": "geometry",
	"stylers": [
	  {
	    "color": "#c9c9c9"
	  }
	]
      },
      {
	"featureType": "water",
	"elementType": "labels.text.fill",
	"stylers": [
	  {
	    "color": "#9e9e9e"
	  }
	]
      }
    ],
    center: {lat: 40.759234, lng: -74.0147407},
    zoom: 14
  });

  fetchPositions();
  setInterval(fetchPositions, 5 * 1000);
}
