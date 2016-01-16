$State = {
	isText: false,
	wordTime: 950,
	// Time to display a word
	wordAnim: 150,
	// Time to animate a word
	randomInterval: 18000,
	lastRandomIndex: -1,
	randomTimer: null,
	lastMouseUp: -1
};

// From Stack Overflow
// http://stackoverflow.com/questions/1582534/calculating-text-width-with-jquery
$.fn.textWidth = function() {
	var html_org = $(this).html();
	var html_calc = '<span>' + html_org + '</span>';
	$(this).html(html_calc);
	var width = $(this).find('span:first').width();
	$(this).html(html_org);
	return width;
};

// http://stackoverflow.com/questions/19491336/get-url-parameter-jquery
function getUrlParameter(sParam) {
	//var sPageURL = window.location.search.substring(1);
	//alert(sPageURL);
	var sPageURL = 'msg=HI%20I%20AM%20SAMARITAN';
	var sURLVariables = sPageURL.split('&');
	for (var i = 0; i < sURLVariables.length; i++) {
		var sParameterName = sURLVariables[i].split('=');
		if (sParameterName[0] == sParam) {
			return sParameterName[1];
		}
	}
}

$(document).ready(function() {
	// Cache the jquery things
	$State.triangle = $('#triangle');
	$State.text = $('#main p');
	$State.line = $('#main hr');

	$('#userText').addClass('hidden');
	//    $('#userButton').addClass('hidden');
	// Start the triangle blinking
	blinkTriangle();
	chatWebsocket();

	// URL parameter message
	var urlMsg = getUrlParameter('msg');
	urlMsg = urlMsg.split('%20').join(' ').split('%22').join('').split('%27').join("'");
	$State.phraselist = [urlMsg];

	setTimeout(function() {
		executeSamaritan(urlMsg);
	},
	$State.wordTime);

});

var blinkTriangle = function() {
	// Stop blinking if samaritan is in action
	if ($State.isText) return;
	$State.triangle.fadeTo(500, 0).fadeTo(500, 1, blinkTriangle);
};

var chatWebsocket = function() {
	var $msg = $('#msg');
	var $text = $('#userText');

	var WebSocket = window.WebSocket || window.MozWebSocket;

	if (WebSocket) {

		try {
			var socket = new WebSocket('wss://samaritan.tech:8443/websocket');
		} catch(e) {}
	}

	if (!socket) {
		socket.onmessage = function(event) {
			//$msg.append('<p>' + event.data + '</p>');
			$State.text.addClass('talk');
			//var hrWidth;
			document.getElementById("msg").innerHTML = event.data;
			if ($State.text.textWidth() < 1) {
				hrWidth = 80;
			} else {
				hrWidth = $State.text.textWidth() + 20
			}
			$State.line.animate({
				'width': (hrWidth) + "px"
			},
			{
				'duration': $State.wordAnim
			})
		};

		$('form').submit(function() {
			socket.send($text.val());
			$text.val('').select();
			return false;
		});
	} else {
		var error_sleep_time = 500;

		function poll() {
			$.ajax({
				url: 'https://samaritan.tech:8443/ajax',
				type: 'GET',
				success: function(event) {
					//$msg.append('<p>' + event + '</p>');
					$State.text.addClass('talk');
					//var hrWidth;
					document.getElementById("msg").innerHTML = event;
					if ($State.text.textWidth() < 1) {
						hrWidth = 80;
					} else {
						hrWidth = $State.text.textWidth() + 20
					}
					$State.line.animate({
						'width': (hrWidth) + "px"
					},
					{
						'duration': $State.wordAnim
					})
					console.log(event);
					error_sleep_time = 500;
					poll();
				},
				error: function() {
					error_sleep_time *= 2;
					setTimeout(poll, error_sleep_time);
				}
			});
		}

		poll();

		$('form').submit(function() {
			$.ajax({
				url: 'https://samaritan.tech:8443/ajax',
				type: 'POST',
				data: {
					text: $text.val()
				},
				success: function() {
					$text.val('').select();
				}
			});
			return false;
		});
	}
};

var executeSamaritan = function(phrase) {
	if ($State.isText) return;

	$State.isText = true;
	var phraseArray = phrase.split(" ");
	// First, finish() the blink animation and
	// scale down the marker triangle
	$State.triangle.finish().animate({
		'font-size': '0em',
		'opacity': '1'
	},
	{
		'duration': $State.wordAnim,
		// Once animation triangle scale down is complete...
		'done': function() {
			var timeStart = 0;
			// Create timers for each word
			phraseArray.forEach(function(word, i) {
				var wordTime = $State.wordTime;
				if (word.length > 8) wordTime *= (word.length / 8);
				setTimeout(function() {
					// Set the text to black, and put in the word
					// so that the length can be measured
					$State.text.addClass('hidden').html(word);
					// Then animate the line with extra padding
					$State.line.animate({
						'width': ($State.text.textWidth() + 20) + "px"
					},
					{
						'duration': $State.wordAnim,
						// When line starts animating, set text to white again
						'start': $State.text.removeClass('hidden')

					})
				},
				(timeStart + $State.wordAnim));
				timeStart += wordTime;
			});

			// Set a final timer to hide text and show triangle
			setTimeout(function() {
				// Clear the text
				$State.text.html("");
				// Animate trinagle back in
				$State.triangle.finish().animate({
					'font-size': '4em',
					'opacity': '1'
				},
				{
					'duration': $State.wordAnim,
					// Once complete, blink the triangle again and animate the line to original size
					'done': function() {
						$State.isText = false;
						blinkTriangle();
						// show textArea
						$('#userText').removeClass('hidden');
						document.querySelector('input').focus();

						//                            $('#userButton').removeClass('hidden');
						$State.line.animate({
							'width': "80px"
						},
						{
							'duration': $State.wordAnim,
							'start': $State.text.removeClass('hidden')
						})
					}
				});
			},
			timeStart + $State.wordTime);
		}
	});
};