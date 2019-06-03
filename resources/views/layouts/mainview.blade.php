<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <link rel="stylesheet" href="{{ URL::asset('css/navigator.css') }}">
    <link rel="stylesheet" href="{{ URL::asset('css/main.css') }}">
    @yield('extension_files')
    @yield('extension_metas')
</head>
<body>
    <div class="navigator">
		<ul>
			<li><a href="http://103.114.161.226">Home</a></li>
			<li><a href="http://103.114.161.226/articles">Article</a></li>
			<li><a href="https://github.com/Myriad-Dreamin">Code</a></li>
		</ul>
        <div class="clear"></div>
    </div>
    
    <div>
    @yield('main_content')
    </div>
</body>
</html>