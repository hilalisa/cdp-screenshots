<?php
// DO NOT EDIT! Generated by Protobuf-PHP protoc plugin 1.0
// Source: query.proto

namespace Vitess\Proto\Query {

  class CommitResponse extends \DrSlump\Protobuf\Message {


    /** @var \Closure[] */
    protected static $__extensions = array();

    public static function descriptor()
    {
      $descriptor = new \DrSlump\Protobuf\Descriptor(__CLASS__, 'query.CommitResponse');

      foreach (self::$__extensions as $cb) {
        $descriptor->addField($cb(), true);
      }

      return $descriptor;
    }
  }
}

