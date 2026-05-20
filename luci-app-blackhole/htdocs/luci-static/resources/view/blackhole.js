'use strict';
'require view';
'require form';
'require uci';
'require fs';
'require ui';
'require rpc';

var callInitAction = rpc.declare({
	object: 'rc',
	method: 'init',
	params: [ 'name', 'action' ],
	expect: { result: false }
});

return view.extend({
	load: function() {
		return Promise.all([
			uci.load('blackhole')
		]);
	},
	render: function(data) {
		var m, s, o;

		m = new form.Map('blackhole', _('Blackhole Webserver'), _('Configure the lightweight blackhole server that captures, logs, and mirrors HTTP requests.'));

		s = m.section(form.TypedSection, 'blackhole');
		s.anonymous = true;
		s.addremove = false;

		s.tab('general', _('General Settings'));
		s.tab('pixel', _('Tracking Pixel'));
		s.tab('logs', _('Logs & Log Viewer'));
		s.tab('files', _('File Browser & Content Editor'));
		s.tab('actions', _('Service Actions'));

		// --- General Tab ---
		o = s.taboption('general', form.Flag, 'enable', _('Enable Server'));
		o.rmempty = false;

		o = s.taboption('general', form.Value, '_bind', _('Listen Address (IP:Port)'));
		o.default = '0.0.0.0:8080';
		o.cfgvalue = function(section_id) {
			var ip = uci.get('blackhole', section_id, 'ip') || '0.0.0.0';
			var port = uci.get('blackhole', section_id, 'port') || '8080';
			return ip + ':' + port;
		};
		o.write = function(section_id, formvalue) {
			var parts = formvalue.split(':');
			if (parts.length >= 2) {
				uci.set('blackhole', section_id, 'ip', parts[0]);
				uci.set('blackhole', section_id, 'port', parts[1]);
			} else {
				uci.set('blackhole', section_id, 'port', formvalue);
			}
		};

		o = s.taboption('general', form.Value, 'root', _('Mirroring Directory (Root)'));
		o.default = '/mnt/largedata/blackholeserver/mirrored';

		o = s.taboption('general', form.Value, 'content', _('Content Directory'));
		o.default = '/mnt/largedata/blackholeserver/content';

		o = s.taboption('general', form.Value, 'log', _('Log File Path'));
		o.default = '/mnt/largedata/blackholeserver/blackhole.log';


		// --- Pixel Tab ---
		o = s.taboption('pixel', form.Flag, 'pixel_enable', _('Enable Tracking Pixel'));
		o.default = '1';

		o = s.taboption('pixel', form.Value, 'pixel_file', _('Custom Pixel File'), _('Absolute path to a custom image on the router.'));
		o.depends('pixel_enable', '1');

		o = s.taboption('pixel', form.Value, 'pixel_hex', _('Custom Pixel Hex'), _('Raw hex string for tracking pixel. Used if Custom File is empty.'));
		o.depends('pixel_enable', '1');


		// --- Logs Tab ---
		o = s.taboption('logs', form.ListValue, 'log_level', _('Log Level'));
		o.value('debug', _('Debug'));
		o.value('info', _('Info'));
		o.value('error', _('Error'));
		o.default = 'info';

		o = s.taboption('logs', form.Value, 'log_max_size', _('Max Log Size (MB)'));
		o.default = '10';

		o = s.taboption('logs', form.DummyValue, '_logview', _('Live Log Viewer'));
		o.rawhtml = true;
		o.cfgvalue = function(section_id) {
			var logPath = uci.get('blackhole', section_id, 'log') || '/var/log/blackhole.log';
			return fs.read_direct(logPath).then(function(content) {
				return '<pre style="max-height: 400px; overflow-y: scroll; background: #333; color: #fff; padding: 10px;">' + (content || 'Log file is empty or missing.') + '</pre>';
			}).catch(function() {
				return '<pre style="max-height: 400px; background: #333; color: red; padding: 10px;">Failed to read log file.</pre>';
			});
		};


		// --- Files Tab ---
		o = s.taboption('files', form.DummyValue, '_files_browser');
		o.rawhtml = true;
		o.cfgvalue = function(section_id) {
			var rootDir = uci.get('blackhole', section_id, 'root') || '/tmp/blackhole_root';
			var contentDir = uci.get('blackhole', section_id, 'content') || '/tmp/blackhole_content';
			
			return Promise.all([
				fs.list(rootDir).catch(function(){ return []; }),
				fs.list(contentDir).catch(function(){ return []; })
			]).then(function(results) {
				var mirrored = results[0];
				var content = results[1];
				var html = '<div style="display:flex; gap: 20px;">';
				
				html += '<div style="flex:1; border: 1px solid #ccc; padding: 10px;"><h3>Mirrored (' + rootDir + ')</h3><ul>';
				if (mirrored.length === 0) html += '<li><i>Empty</i></li>';
				for (var i=0; i<mirrored.length; i++) html += '<li>' + (mirrored[i].type === 'directory' ? '📁 ' : '📄 ') + mirrored[i].name + '</li>';
				html += '</ul></div>';
				
				html += '<div style="flex:1; border: 1px solid #ccc; padding: 10px;"><h3>Content (' + contentDir + ')</h3><ul>';
				if (content.length === 0) html += '<li><i>Empty</i></li>';
				for (var j=0; j<content.length; j++) html += '<li>' + (content[j].type === 'directory' ? '📁 ' : '📄 ') + content[j].name + '</li>';
				html += '</ul></div>';
				
				html += '</div>';
				return html;
			});
		};

		o = s.taboption('files', form.DummyValue, '_spacer', '');
		o.rawhtml = true;
		o.cfgvalue = function() { return '<hr/><h3>Create/Edit Content File</h3>'; };

		o = s.taboption('files', form.Value, '_new_filename', _('File Name'), _('Relative to Content Directory (e.g. index.html)'));
		
		o = s.taboption('files', form.TextValue, '_new_content', _('File Content'));
		o.rows = 10;
		
		o = s.taboption('files', form.Button, '_save_file', _('Save File'));
		o.inputstyle = 'apply';
		o.onclick = function(ev, section_id) {
			var contentDir = uci.get('blackhole', section_id, 'content') || '/tmp/blackhole_content';
			var elFilename = document.getElementById('cbid.blackhole.' + section_id + '._new_filename');
			var elContent = document.getElementById('cbid.blackhole.' + section_id + '._new_content');
			
			var filename = elFilename ? elFilename.value : '';
			var content = elContent ? elContent.value : '';
			
			if (!filename) {
				ui.addNotification(null, E('p', _('Please specify a filename.')));
				return;
			}
			
			var fullPath = contentDir + '/' + filename;
			return fs.write(fullPath, content || '').then(function() {
				ui.addNotification(null, E('p', _('File saved to ' + fullPath)));
				if(elFilename) elFilename.value = '';
				if(elContent) elContent.value = '';
			}).catch(function(e) {
				ui.addNotification(null, E('p', _('Failed to save file: ' + e.message)));
			});
		};


		// --- Actions Tab ---
		o = s.taboption('actions', form.Button, '_start', _('Start Webserver'));
		o.inputstyle = 'apply';
		o.onclick = function() {
			return callInitAction('blackhole', 'start').then(function() {
				ui.addNotification(null, E('p', _('Sent start command to Blackhole.')));
			});
		};

		o = s.taboption('actions', form.Button, '_stop', _('Stop Webserver'));
		o.inputstyle = 'reset';
		o.onclick = function() {
			return callInitAction('blackhole', 'stop').then(function() {
				ui.addNotification(null, E('p', _('Sent stop command to Blackhole.')));
			});
		};

		o = s.taboption('actions', form.Button, '_restart', _('Restart Webserver'));
		o.inputstyle = 'apply';
		o.onclick = function() {
			return callInitAction('blackhole', 'restart').then(function() {
				ui.addNotification(null, E('p', _('Sent restart command to Blackhole.')));
			});
		};

		return m.render();
	}
});
