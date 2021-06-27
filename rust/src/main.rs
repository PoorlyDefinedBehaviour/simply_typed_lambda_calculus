use std::collections::HashMap;

#[derive(Debug, Clone, PartialEq)]
enum Type {
  Int,
  Arrow {
    parameter_type: Box<Type>,
    body_type: Box<Type>,
  },
}

type Context<Key, Value> = HashMap<Key, Value>;

#[derive(Debug, Clone)]
enum Expression {
  Int(i64),
  Variable(String),
  Abstraction {
    parameter: String,
    parameter_type: Type,
    body: Box<Expression>,
  },
  Application {
    argument: Box<Expression>,
    function: Box<Expression>,
  },
}

fn infer(context: &mut Context<String, Type>, expression: &Expression) -> Result<Type, String> {
  use Expression::*;

  match expression {
    Int(_) => Ok(Type::Int),
    Variable(variable_name) => match context.get(variable_name) {
      None => Err(format!("undefined variable {}", variable_name)),
      Some(variable_type) => Ok(variable_type.clone()),
    },
    Abstraction {
      parameter,
      parameter_type,
      body,
    } => {
      context.insert(parameter.clone(), parameter_type.clone());

      dbg!(&parameter, &parameter_type, &body);

      let body_type = infer(context, body)?;

      Ok(Type::Arrow {
        parameter_type: Box::new(parameter_type.clone()),
        body_type: Box::new(body_type),
      })
    }
    Application { argument, function } => {
      let argument_type = infer(context, argument)?;
      let function_type = infer(context, function)?;

      match function_type {
        Type::Arrow {
          parameter_type,
          body_type,
        } => {
          if argument_type != *parameter_type {
            Err(format!(
              "expected type {:?}, got {:?}",
              parameter_type, argument_type
            ))
          } else {
            Ok(*body_type)
          }
        }
        _ => Err(format!("unknown expression {:?}", expression)),
      }
    }
  }
}

fn main() {
  use Expression::*;

  let identity = Abstraction {
    parameter: "a".to_owned(),
    parameter_type: Type::Int,
    body: Box::new(Variable("a".to_owned())),
  };

  let expression = Application {
    argument: Box::new(Int(10)),
    function: Box::new(identity),
  };

  let mut typing_context = Context::new();

  dbg!(infer(&mut typing_context, &expression));
}
