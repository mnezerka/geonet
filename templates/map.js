const clusterColors = [
    'Blue',
    'Green',
    'Lime',
    'Yellow',
    'Orange',
    'Pink',
    'Violet',
    'DarkOrange',
    'Gold'
]

//////////////////////////////////////// global vars

// one layer for each highlighted track
var trackLayers = {}

//////////////////////////////////////// checkbox

// check box control
function createCheckBox(parentElement, trackId) {
    // determine if checkbox should be checked or not
    let checked = trackLayers[trackId] !== undefined;

    var newCheckBox = document.createElement('input');
    newCheckBox.setAttribute('type', 'checkbox');
    newCheckBox.checked = checked;
    newCheckBox.setAttribute('data-track-id', trackId)
    newCheckBox.setAttribute('onclick', 'onCheckBoxClick(this)')
    parentElement.appendChild(newCheckBox);
}

function onCheckBoxClick(cb){

    let trackId = Number(cb.getAttribute('data-track-id'));
    if (trackId) {

        if (cb.checked) {
            highlightedAdd(trackId);
        } else {
            highlightedRemove(trackId);
        }
    }
};

//////////////////////////////////////// highlighting

function isFeatureOnTrack(feature, trackId) {

    // check if geojson object has relevant attributes
    if (feature.type != 'Feature') {  return false; }
    if (!feature.properties || !feature.properties.tracks) {
        return false
    }

    return feature.properties.tracks.indexOf(trackId) > -1
}

function highlightedAdd(trackId) {

    // if layer for track already exists, do nothing
    if (trackLayers[trackId]) {
        return;
    }

    let l = L.geoJson(
        geojson,
        {
            style: styleFuncHighlighted,
            filter: function(f) { return isFeatureOnTrack(f, trackId) },
            pointToLayer: pointToLayerFunc,
            onEachFeature: onEachFeatureFunc
        }
    ).addTo(map);

    trackLayers[trackId] = {
        id: trackId,
        layer: l
    }
}

function highlightedRemove(trackId) {

    // if layer for track doesn't exist, do nothing
    if (!trackLayers[trackId]) {
        return;
    }

    map.removeLayer(trackLayers[trackId].layer)
    delete(trackLayers[trackId])
}



//////////////////////////////////////// something


function track_id_to_track_name(id) {
    if (meta.tracks !== undefined) {
        for (let i = 0; i < meta.tracks.length; i++) {
            let t = meta.tracks[i]
            if (t.id == id) {
                let result = t.id + ":";
                if (t.meta.post_url) { result += '<a href="' + t.meta.post_url + '">'; }
                result += t.meta.post_title
                if (t.meta.track_title) {
                    result += ' (' + t.meta.track_title + ')';
                }
                if (t.meta.post_url) { result += '</a>'; }
                return result
                break;
            }
        }
    }
    return 'track: ' + id
}

// enhance tracks of colors, titles, etc.
if (meta.tracks !== undefined) {
    for (let i = 0; i < meta.tracks.length; i++) {
        meta.tracks[i].color = clusterColors[i % clusterColors.length]
    }
}

function componentToHex(c) {
    var hex = c.toString(16);
    return hex.length == 1 ? "0" + hex : hex;
}

const map = L.map('map').setView([51.505, -0.09], 13);

const tiles = L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
}).addTo(map);

L.control.scale().addTo(map);

const geojsonMarkerOptions = {
    radius: 5,
    fillColor: "#fff969",
    color: "#cfc939",
    weight: 1,
    opacity: 1,
    fillOpacity: 0.8
};

const getColorForTracks = function(tracks) {

    if (!useTrackColors) {
        return 'red'
    }

    if (tracks.length == 0) {
        return
    }

    let first = tracks[0]
    if (meta.tracks) {
        for (let i = 0; i < meta.tracks.length; i++) {
            if (meta.tracks[i].id == first) {
                return meta.tracks[i].color
            }
        }
    }
}

const styleFuncHighlighted = function(feature) {

    if (feature.geometry.type === 'LineString') {
        let tstyle = {
            color: 'blue'
        }
        return tstyle
    }

    if (feature.geometry.type === 'Point') {
        let tstyle = {
            fillColor: 'blue',
            radius: 5,
        }

        if (feature.properties.crossing === true) {
            tstyle.fillColor = 'blue'
            tstyle.radius = 10 
         }

        return tstyle
    }

    return {}
}


// executed for each geojson feature (point, linestring, etc...)
const styleFunc = function(feature) {

    if (feature.geometry.type === 'LineString') {
        let tstyle = {
            color: getColorForTracks(feature.properties.tracks),
        }
        return tstyle
    }

    if (feature.geometry.type === 'Point') {
        let tstyle = {
            fillColor: getColorForTracks(feature.properties.tracks),
            radius: 5
        }

        if (feature.properties.crossing === true) {
            tstyle.fillColor = 'red'
            tstyle.radius = 10 
         }

        return tstyle
    }

    return {}
}

// this is executed only for points
const pointToLayerFunc = function(feature, latlng) {
    opts = {}
    Object.assign(opts, geojsonMarkerOptions);
    return L.circleMarker(latlng, opts);
}

// this is executed for each geojson feature (point, linestring, etc.)
const onEachFeatureFunc = function (feature, layer) {
    // dynamic popup sinc popup content has some js logic that must
    // be evaluated at time of opening
    layer.on({
        click: function(e) { openFeaturePopup(feature, layer); },
        popupclose: function(e) { layer.unbindPopup(); }
    });
}

function openFeaturePopup(feature, layer) {
    let el = document.createElement('div');

    // title
    if (feature.properties.id !== undefined) {
        let elTitle = document.createElement('p');
        elTitle.innerHTML = feature.geometry.type + ' id: ' + feature.properties.id
        el.appendChild(elTitle)
    }

    // tracks
    if (feature.properties.tracks !== undefined) {

        let elTracks = document.createElement('div');

        let tracks = feature.properties.tracks;
        for (let i = 0; i < tracks.length; i++) {

            let elTrack = document.createElement('div');
            elTrack.setAttribute('class', 'track-info');

            // checkbox
            createCheckBox(elTrack, tracks[i])

            // description
            let elTrackDesc = document.createElement('span');
            elTrackDesc.innerHTML = track_id_to_track_name(tracks[i])
            elTrack.appendChild(elTrackDesc);

            elTracks.appendChild(elTrack);
        }
        el.appendChild(elTracks);
    }

    layer.bindPopup(el).openPopup();
}


matchesLayer = L.geoJson(
    geojson,
    {
        style: styleFunc,
        pointToLayer: pointToLayerFunc,
        onEachFeature: onEachFeatureFunc
    }).addTo(map);

map.fitBounds(matchesLayer.getBounds())
