/* global d3, google */
'use strict';

let map;
let data;
let infoWindow;

const PATH = `M20 21c-1.39 0-2.78-.47-4-1.32-2.44 1.71-5.56 1.71-8 0C6.78 20.53 5.39 21 4 21H2v2h2c1.38 0 2.74-.35 4-.99 2.52 1.29 5.48 1.29 8 0 1.26.65 2.62.99 4 .99h2v-2h-2zM3.95 19H4c1.6 0 3.02-.88 4-2 .98 1.12 2.4 2 4 2s3.02-.88 4-2c.98 1.12 2.4 2 4 2h.05l1.89-6.68c.08-.26.06-.54-.06-.78s-.34-.42-.6-.5L20 10.62V6c0-1.1-.9-2-2-2h-3V1H9v3H6c-1.1 0-2 .9-2 2v4.62l-1.29.42c-.26.08-.48.26-.6.5s-.15.52-.06.78L3.95 19zM6 6h12v3.97L12 8 6 9.97V6z`;

const colorScale = d3.scaleLinear()
      .domain([0, 60000])
      .range([ '#0097DE', '#000'])
      .clamp(true);

function geoJsonFromResponse(response) {
  const geojson = {
    type: 'FeatureCollection',
    features: [],
  };

  for (let i = 0; i < response.length; i++) {
    const position = response[i];
    const lat = position.PositionReport.Lat
    const lng = position.PositionReport.Lon
    const utcTime = (new Date()).getTime();
    const feature = {
      type: 'Feature',
      geometry: {
	type: 'Point',
	coordinates: [lng, lat],
      },
      properties: {
	mmsi: position.PositionReport.MMSI,
	position_report: position.PositionReport,
	age: utcTime - Date.parse(position.Timestamp),
      },
    };

    console.log(position.Timestamp);
    console.log(Date.parse(position.Timestamp));
    console.log(utcTime);
    console.log(feature.properties.age)
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

    data.setStyle(feature => {
      // Use a path that outlines the shape of a pin.
      return {
	icon: {
	  path: PATH,
	  fillColor: colorScale(feature.getProperty('age')),
	  fillOpacity: 1,
	  strokeColor: '#FFF',
	  strokeWeight: 0.5,
	  anchor: new google.maps.Point(12,12),
	}
      };
    });

    data.addListener('click', event => {
      if (infoWindow) {
	infoWindow.close();
      }
      const mmsi = event.feature.getProperty('mmsi');
      const position_report = JSON.stringify(event.feature.getProperty('position_report'));

      infoWindow = new google.maps.InfoWindow({
	content: `<div><h3>${mmsi}</h3><p>${position_report}</p></div>`,
	position: event.feature.getGeometry().get(0),
      });
      infoWindow.open(map, data);
    });

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

  // Close the info window on outside clicks.
  map.addListener('click', () => {
    if (infoWindow) {
      infoWindow.close();
    }
  });

  fetchPositions();
  setInterval(fetchPositions, 5 * 1000);
}
