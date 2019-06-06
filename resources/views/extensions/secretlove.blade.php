@extends('layouts.mainview')

@section('extension_files')
    <link rel="stylesheet" href="{{ URL::asset('css/extension/secretlove.css') }}">
@stop

@section('extension_metas')
    <title>Marshomallo</title>
@stop

@section('main_content')
    <div class="index_box">
        <div class="title">Send Marshmello to me</div>
        {!! Form::open() !!}
        <div class="margin-left-5 margin-right-5">
            <div>
                {!! Form::text('RequestName',null,['class'=>'request-name', 'placeholder'=>'Your Name']) !!}
                
            </div>
            <div class="clear"></div>
            <div>
                {!! Form::text('MarshmelloName',null,['class'=>'marshmello-name', 'placeholder'=>'Title']) !!}
            </div>
            <div>
                {!! Form::textarea('content',null,['class'=>'honey']) !!}
            </div>
            <div>
                {!! Form::submit('Send',['class'=>'sendbtn']) !!}
            </div>
        </div>
        {!! Form::close() !!}
    </div>
@stop
